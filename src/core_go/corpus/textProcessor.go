package corpus

import (
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sync"

	"github.com/tcc2-davi-arthur/utils"
)

const (
	dirClen = "./misc/corpus/clean"
	dirTxt  = "./misc/corpus/txt"
	dir     = "./misc/corpus/pdf"
)

func TextProcessor(maxWorkers int) {

	err := errors.Join(
		os.MkdirAll(dir, 0755),
		os.MkdirAll(dirTxt, 0755),
		os.MkdirAll(dirClen, 0755),
	)
	if err != nil {
		log.Fatal(err)
	}

	files, err := os.ReadDir(dir)
	if err != nil {
		log.Fatal(err)
	}

	var wgCleaner sync.WaitGroup
	workerLimit := make(chan struct{}, maxWorkers)

	for _, f := range files {
		if f.IsDir() || filepath.Ext(f.Name()) != ".pdf" {
			continue
		}

		wgCleaner.Add(1)
		workerLimit <- struct{}{}

		go func(name string) {
			defer func() {
				<-workerLimit
			}()
			processPDF(filepath.Join(dir, name), &wgCleaner)
		}(f.Name())
	}

	wgCleaner.Wait()
}

func processPDF(path string, wg *sync.WaitGroup) {
	defer wg.Done()

	// Extrai texto
	out, err := exec.Command("pdftotext", path, "-").Output()
	if err != nil {
		log.Printf("Erro pdftotext %s: %v\n", path, err)
		return
	}

	// Salva texto convertido
	filename := filepath.Base(path)[:len(filepath.Base(path))-len(filepath.Ext(path))]
	txtPath := fmt.Sprintf("./misc/corpus/txt/%s.txt", filename)
	if err := os.WriteFile(txtPath, out, 0744); err != nil {
		log.Printf("Erro salvando %s: %v\n", txtPath, err)
		return
	}

	// Limpa o texto
	clean, err := utils.CleanText(string(out))
	if err != nil {
		log.Printf("Erro limpando %s: %v\n", path, err)
		return
	}

	// Salva texto limpo
	cleanPath := fmt.Sprintf("./misc/corpus/clean/%s_clean.txt", filename)
	if err := os.WriteFile(cleanPath, []byte(clean), 0744); err != nil {
		log.Printf("Erro salvando %s: %v\n", cleanPath, err)
		return
	}

	log.Printf("Processado: %s\n", path)
}
