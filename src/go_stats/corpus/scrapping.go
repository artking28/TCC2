package corpus

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
	pdfcpuapi "github.com/pdfcpu/pdfcpu/pkg/api"
)

const (
	apiURL    = "https://www.camara.leg.br/busca-api/api/v1/busca/proposicoes/_search"
	tempDir   = "./misc/corpus/temp"
	corpusDir = "./misc/corpus/pdf"
)

var (
	totalPages int
	totalDocs  int
	mu         sync.Mutex
	wgScrap    sync.WaitGroup
)

type apiResponse struct {
	Hits struct {
		Hits []struct {
			ID string `json:"_id"`
		} `json:"hits"`
	} `json:"hits"`
}

type searchBody struct {
	Order             string `json:"order"`
	Pagina            int    `json:"pagina"`
	Q                 string `json:"q"`
	TiposDeProposicao string `json:"tiposDeProposicao"`
}

func StartScrapping(maxPages, maxWorkers int, delayPerReq time.Duration) {

	err := os.MkdirAll(tempDir, 0755)
	if err != nil {
		log.Fatal(err)
	}

	err = os.MkdirAll(corpusDir, 0755)
	if err != nil {
		log.Fatal(err)
	}

	tasks := make(chan string, 200)
	for i := 0; i < maxWorkers; i++ {
		wgScrap.Add(1)
		go worker(maxPages, delayPerReq, tasks)
	}

pageLoop:
	for page := 1; ; page++ {
		if getTotalPages() >= maxPages {
			break
		}

		body := searchBody{
			Order:             "relevancia",
			Pagina:            page,
			Q:                 "*",
			TiposDeProposicao: "PEC,PLP,PL,PDL",
		}

		data, _ := json.Marshal(body)
		req, _ := http.NewRequest("POST", apiURL, strings.NewReader(string(data)))
		req.Header.Set("Content-Type", "application/json")

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			log.Println("Erro na API:", err)
			break
		}

		var apiResp apiResponse
		respBody, _ := io.ReadAll(resp.Body)
		resp.Body.Close()

		if err := json.Unmarshal(respBody, &apiResp); err != nil {
			log.Println("Erro ao parsear JSON:", err)
			break
		}

		if len(apiResp.Hits.Hits) == 0 {
			fmt.Println("Sem mais resultados")
			break
		}

		for _, h := range apiResp.Hits.Hits {
			if getTotalPages() >= maxPages {
				break pageLoop
			}
			link := fmt.Sprintf("https://www.camara.leg.br/proposicoesWeb/fichadetramitacao?idProposicao=%s", h.ID)
			tasks <- link
		}

		time.Sleep(delayPerReq)
	}

	close(tasks)
	wgScrap.Wait()

	fmt.Printf("Finalizado: %d PDFs válidos, %d páginas totais\n", getTotalDocs(), getTotalPages())
}

func worker(maxPages int, delayPerReq time.Duration, tasks <-chan string) {
	defer wgScrap.Done()
	for url := range tasks {
		if getTotalPages() >= maxPages {
			return
		}

		pdfURL, err := getPDFLink(url)
		if err != nil {
			log.Printf("[ERRO] Link PDF: %s | %v\n", url, err)
			continue
		}
		pdfURL = strings.Split(pdfURL, "&filename")[0]

		tempFile := fmt.Sprintf("%s/temp_%d.pdf", tempDir, time.Now().UnixNano())

		err = baixarPdfCamara(tempFile, pdfURL)
		if err != nil {
			log.Printf("[ERRO] Download PDF: %s | %v\n", pdfURL, err)
			continue
		}

		n, err := pdfcpuapi.PageCountFile(tempFile)
		if err != nil {
			os.Remove(tempFile)
			continue
		}

		if n >= 1 {
			newTotal := addPages(n)
			if newTotal > maxPages {
				os.Remove(tempFile)
				return
			}

			id := incDocs()
			finalFile := fmt.Sprintf("%s/doc_%04d.pdf", corpusDir, id)
			os.Rename(tempFile, finalFile)
			log.Printf("[OK] PDF salvo (%d pág): %s | Total: %d páginas\n", n, finalFile, getTotalPages())
		} else {
			os.Remove(tempFile)
		}

		time.Sleep(delayPerReq)
	}
}

func getPDFLink(detailURL string) (string, error) {
	resp, err := http.Get(detailURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return "", err
	}

	link, exists := doc.Find("a.linkDownloadTeor").First().Attr("href")
	if !exists {
		return "", fmt.Errorf("link PDF não encontrado")
	}

	if !strings.HasPrefix(link, "http") {
		link = "https://www.camara.leg.br" + link
	}
	return link, nil
}

func baixarPdfCamara(fileName, url string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	return os.WriteFile(fileName, data, 0644)
}

// ---- Contadores protegidos ----

func getTotalPages() int {
	mu.Lock()
	defer mu.Unlock()
	return totalPages
}

func addPages(n int) int {
	mu.Lock()
	defer mu.Unlock()
	totalPages += n
	return totalPages
}

func getTotalDocs() int {
	mu.Lock()
	defer mu.Unlock()
	return totalDocs
}

func incDocs() int {
	mu.Lock()
	defer mu.Unlock()
	totalDocs++
	return totalDocs
}
