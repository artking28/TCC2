package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/tcc2-davi-arthur/utils"
)

const (
	DylibPath     = "/Users/arthurandrade/Desktop/artking28/onnxruntime/build/MacOS/Release/Release/libonnxruntime.dylib"
	InputsPath    = "/Users/arthurandrade/Desktop/SENAC/Tcc2/misc/searchLegalInputs.json"
	TokenizerPath = "/Users/arthurandrade/Desktop/SENAC/Tcc2/misc/bert/tokenizer.json"
	LawsFolder    = "/Users/arthurandrade/Desktop/SENAC/Tcc2/misc/corpus/clean"
	OnnxPath      = "/Users/arthurandrade/Desktop/SENAC/Tcc2/misc/bert/model.onnx"
	OutputPath    = "output.json"
)

type (
	SearchInputs map[string][]InputSample

	InputSample struct {
		Input string `json:"input"`
		Bert  []int  `json:"bert,omitempty"`
		BertT int64  `json:"bertT,omitempty"`
	}

	Result struct {
		Index int
		Score float32
	}
)

func search(queryEmb []float32, docEmbeddings [][]float32) ([]int, int64) {
	start := time.Now()
	results := make([]Result, len(docEmbeddings))

	for i, docEmb := range docEmbeddings {
		score := utils.CosineSimVecs(queryEmb, docEmb)
		results[i] = Result{Index: i, Score: score}
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	finalIndices := make([]int, len(results))
	for i, res := range results {
		finalIndices[i] = res.Index + 1
	}
	return finalIndices, time.Since(start).Microseconds()
}

func loadTexts(folder string) ([]string, []string, error) {
	var filenames, texts []string
	files, err := os.ReadDir(folder)
	if err != nil {
		return nil, nil, err
	}
	for _, f := range files {
		if filepath.Ext(f.Name()) == ".txt" {
			path := filepath.Join(folder, f.Name())
			content, err := os.ReadFile(path)
			if err != nil {
				continue
			}
			filenames = append(filenames, f.Name())
			texts = append(texts, string(content))
		}
	}
	return filenames, texts, nil
}

func main() {
	log.Println("Inicializando ONNX Runtime...")
	if err := utils.InitONNX(DylibPath); err != nil {
		log.Fatal("Falha no InitONNX:", err)
	}
	defer utils.DestroyONNX()

	log.Println("Carregando BERT (Static Tensors)...")
	bert, err := utils.LoadBert(OnnxPath, TokenizerPath)
	if err != nil {
		log.Fatal("Falha ao carregar BERT:", err)
	}
	defer bert.Close()

	log.Println("Lendo arquivos...")
	files, texts, err := loadTexts(LawsFolder)
	if err != nil {
		log.Fatal(err)
	}
	nDocs := len(files)

	log.Println("Gerando embeddings dos documentos...")
	docEmbeddings := make([][]float32, nDocs)

	var totalDocsTime time.Duration // Acumulador de tempo

	for i, text := range texts {

		// CRONÔMETRO INÍCIO
		start := time.Now()

		emb, err := bert.Apply(text)

		// CRONÔMETRO FIM
		duration := time.Since(start)
		totalDocsTime += duration

		if err != nil {
			log.Fatalf("Erro no doc %d: %v", i, err)
		}
		docEmbeddings[i] = emb

		fmt.Printf("Processados %d/%d...\r", i, nDocs)
	}
	fmt.Println("\nEmbeddings concluídos.")
	log.Println("Lendo inputs...")

	jsonBytes, err := os.ReadFile(InputsPath)
	if err != nil {
		log.Fatal(err)
	}

	var inputs SearchInputs
	if err = json.Unmarshal(jsonBytes, &inputs); err != nil {
		log.Fatal(err)
	}

	finalOutput := make(SearchInputs)
	groups := []string{"words10", "words20", "words40"}

	// Map para guardar o tempo total por grupo
	groupTimes := make(map[string]time.Duration)
	groupCounts := make(map[string]int)

	for _, group := range groups {
		log.Printf("Processando grupo %s...", group)
		samples := inputs[group]
		processedSamples := make([]InputSample, 0, len(samples))

		var groupTotalTime time.Duration // Acumulador do grupo atual

		for _, sample := range samples {

			// CRONÔMETRO INÍCIO (QUERY)
			start := time.Now()

			qEmb, err := bert.Apply(sample.Input)

			// CRONÔMETRO FIM (QUERY)
			duration := time.Since(start)
			groupTotalTime += duration

			if err != nil {
				log.Printf("Erro input: %v", err)
				continue
			}

			ordered, t := search(qEmb, docEmbeddings)
			sample.Bert = ordered
			sample.BertT = t
			processedSamples = append(processedSamples, sample)
		}
		finalOutput[group] = processedSamples

		groupTimes[group] = groupTotalTime
		groupCounts[group] = len(samples)
	}

	outBytes, _ := json.MarshalIndent(finalOutput, "", "  ")
	if err = os.WriteFile(OutputPath, outBytes, 0744); err != nil {
		return
	}

	// --- 6. LOG DE ESTATÍSTICAS FINAIS ---
	log.Println("--- RELATÓRIO DE PERFORMANCE (INFERÊNCIA BERT) ---")

	// Stats Docs
	if nDocs > 0 {
		avgDoc := totalDocsTime / time.Duration(nDocs)
		log.Printf("[Documentos] Total Docs: %d", nDocs)
		log.Printf("[Documentos] Tempo Total: %v", totalDocsTime)
		log.Printf("[Documentos] Média por Doc: %v", avgDoc)
	}

	log.Println("--------------------------------------------------")

	// Stats Inputs por Grupo
	for _, group := range groups {
		count := groupCounts[group]
		if count > 0 {
			total := groupTimes[group]
			avg := total / time.Duration(count)
			log.Printf("[%s] Tempo Total: %v", group, total)
			log.Printf("[%s] Média por Frase: %v", group, avg)
			log.Println("-")
		}
	}

	log.Println("Finalizado!")
}
