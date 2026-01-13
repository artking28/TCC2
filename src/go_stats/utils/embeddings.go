package utils

import (
	"fmt"

	"github.com/sugarme/tokenizer"
	"github.com/sugarme/tokenizer/pretrained"
	ort "github.com/yalue/onnxruntime_go"
)

// --- CONFIGURAÇÃO GLOBAL ---

func InitONNX(dylibPath string) error {
	ort.SetSharedLibraryPath(dylibPath)
	return ort.InitializeEnvironment()
}

func DestroyONNX() {
	_ = ort.DestroyEnvironment()
}

// --- CLIENTE COM TENSORES ESTÁTICOS ---

type BertClient struct {
	Tokenizer *tokenizer.Tokenizer
	Session   *ort.AdvancedSession

	// Mantemos os tensores vivos para reutilizar a memória
	InputTensor *ort.Tensor[int64]
	MaskTensor  *ort.Tensor[int64]
	TypeTensor  *ort.Tensor[int64]
	OutTensor   *ort.Tensor[float32]

	// Constante
	MaxLen int64
}

func LoadBert(onnxPath, tokenizerPath string) (*BertClient, error) {
	// 1. Tokenizer
	tk, err := pretrained.FromFile(tokenizerPath)
	if err != nil {
		return nil, fmt.Errorf("erro tokenizer: %v", err)
	}
	// IMPORTANTE: Definimos Truncation para garantir que nunca passe de 512
	tk.WithTruncation(&tokenizer.TruncationParams{MaxLength: 512})

	// 2. Alocar Tensores ESTÁTICOS (Max Length 512)
	// Como não podemos mudar o tamanho do tensor na sua versão, fixamos em 512.
	maxLen := int64(512)
	shape := ort.NewShape(1, maxLen)

	// Criamos arrays zerados iniciais
	emptyInput := make([]int64, maxLen)

	// Criamos os tensores que viverão para sempre
	inputT, _ := ort.NewTensor(shape, append([]int64(nil), emptyInput...))
	maskT, _ := ort.NewTensor(shape, append([]int64(nil), emptyInput...))
	typeT, _ := ort.NewTensor(shape, append([]int64(nil), emptyInput...))

	// Output tensor (1, 512, 384)
	hiddenSize := int64(384)
	outShape := ort.NewShape(1, maxLen, hiddenSize)
	outT, _ := ort.NewEmptyTensor[float32](outShape)

	// 3. Criar Sessão "Amarrando" esses tensores
	inputNames := []string{"input_ids", "attention_mask", "token_type_ids"}
	outputNames := []string{"last_hidden_state"}

	session, err := ort.NewAdvancedSession(
		onnxPath,
		inputNames,
		outputNames,
		[]ort.Value{inputT, maskT, typeT}, // Inputs fixos
		[]ort.Value{outT},                 // Output fixo
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("erro sessão: %v", err)
	}

	return &BertClient{
		Tokenizer:   tk,
		Session:     session,
		InputTensor: inputT,
		MaskTensor:  maskT,
		TypeTensor:  typeT,
		OutTensor:   outT,
		MaxLen:      maxLen,
	}, nil
}

func (b *BertClient) Close() {
	// Destruir tudo manualmente
	if b.Session != nil {
		b.Session.Destroy()
	}
	if b.InputTensor != nil {
		b.InputTensor.Destroy()
	}
	if b.MaskTensor != nil {
		b.MaskTensor.Destroy()
	}
	if b.TypeTensor != nil {
		b.TypeTensor.Destroy()
	}
	if b.OutTensor != nil {
		b.OutTensor.Destroy()
	}
}

// --- INFERÊNCIA ---

func (b *BertClient) Apply(inputText string) ([]float32, error) {
	// 1. Tokenização
	enc, err := b.Tokenizer.EncodeSingle(inputText)
	if err != nil {
		return nil, err
	}

	realLen := len(enc.Ids)
	if int64(realLen) > b.MaxLen {
		realLen = int(b.MaxLen) // Segurança
	}

	// 2. SOBRESCREVER DADOS DOS TENSORES
	// Pegamos o slice que aponta para a memória do tensor e editamos direto
	inputData := b.InputTensor.GetData()
	maskData := b.MaskTensor.GetData()
	typeData := b.TypeTensor.GetData()

	// Preencher com os dados novos e zerar o resto (Padding)
	for i := 0; i < int(b.MaxLen); i++ {
		if i < realLen {
			inputData[i] = int64(enc.Ids[i])
			maskData[i] = int64(enc.AttentionMask[i])
			typeData[i] = int64(enc.TypeIds[i])
		} else {
			// Padding (Zeros)
			inputData[i] = 0
			maskData[i] = 0 // Importante: Máscara 0 faz o BERT ignorar isso
			typeData[i] = 0
		}
	}

	// 3. Executar (Sem parâmetros, usa os tensores que já atualizamos)
	// OBS: Se sua versão pede options, passe nil
	err = b.Session.Run()
	if err != nil {
		return nil, err
	}

	// 4. Pooling
	// O output terá tamanho (1 * 512 * 384).
	// Precisamos passar o `realLen` para o MeanPooling ignorar o padding.
	rawOutput := b.OutTensor.GetData()

	// Atenção: passamos maskData (que tem zeros no final) e realLen
	// Seu MeanPooling precisa respeitar o maskData para não fazer média com zeros.
	hiddenSize := 384
	pooled := MeanPooling(rawOutput, maskData, int(b.MaxLen), hiddenSize)

	return Normalize(pooled), nil
}
