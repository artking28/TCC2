package main

import (
	"log"
	"testing"

	"github.com/tcc2-davi-arthur/utils"
)

func TestSimples(t *testing.T) {

	// 1. Inicializa Ambiente (Global)
	log.Println("Inicializando ONNX...")
	if err := utils.InitONNX(DylibPath); err != nil {
		t.Fatal("Erro Init:", err)
	}
	defer utils.DestroyONNX()

	// 2. Carrega o Cliente (Modelo + Tokenizer + Tensores Estáticos)
	// Isso é o que faltava no seu exemplo anterior
	log.Println("Carregando Cliente BERT...")
	client, err := utils.LoadBert(OnnxPath, TokenizerPath)
	if err != nil {
		t.Fatal("Erro LoadBert:", err)
	}
	defer client.Close()

	// 3. Executa
	inputText := "Essa é uma frase de teste para gerar embeddings."

	// Note que chamamos client.Apply
	res, err := client.Apply(inputText)
	if err != nil {
		t.Fatal(err)
	}

	log.Println("Sucesso!")
	log.Println("Tamanho do vetor:", len(res)) // Deve ser 384
	log.Println("Primeiros 5 valores:", res[:5])
}
