package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	mgu "github.com/artking28/myGoUtils"
	"github.com/tcc2-davi-arthur/corpus"
	"github.com/tcc2-davi-arthur/models/support"
	"github.com/tcc2-davi-arthur/utils"
	"gorm.io/gorm"
)

// ResultsOutput represents path to the file where benchmark results will be saved.
const ResultsOutput = "./../../misc/resultsT.csv"

// N-gram Size Constants
const (
	Unigram = 1 // Represents a 1-word n-gram.
	Bigram  = 2 // Represents a 2-word n-gram (consecutive words or with jumps).
	Trigram = 3 // Represents a 3-word n-gram (consecutive words or with jumps).
)

// Maximum Jump Limit Constants
const (
	MaxTrigramJumps = 2 // Maximum allowed jumps between words for a Trigram.
	MaxBigramJumps  = 4 // Maximum allowed jumps between words for a Bigram.
)

var csvHeader string

func init() {
	csvHeader = strings.Join([]string{
		"TestID",
		"Algorithm",
		"Pre-Indexed",
		"Normalized jumps",
		"Grams size",
		"Jumps size",
		"Parallel",
		"TotalDocs",
		"TotalTime",
		"AvgSpearmanSim10", "MinSpearmanSim10", "MaxSpearmanSim10", "AvgTime10", "MinTime10", "MaxTime10",
		"AvgSpearmanSim20", "MinSpearmanSim20", "MaxSpearmanSim20", "AvgTime20", "MinTime20", "MaxTime20",
		"AvgSpearmanSim40", "MinSpearmanSim40", "MaxSpearmanSim40", "AvgTime40", "MinTime40", "MaxTime40",
	}, ",") + "\n"
}

// main iterates over all parameter combinations and saves results to a JSON file.
func main() {
	mn, mx, avg := utils.MeasureMemory(mainShift)
	fmt.Printf("Memory usage: \n\tmin: %d\n\tmax: %d\n\tavg: %d\n", mn, mx, avg)
}

func mainShift() {
	//corpus.StartScrapping(5000, 25, 500*time.Millisecond)
	//corpus.TextProcessor(25)

	var id int64 = 1 // Unique counter to identify each test.

	// Defines n-gram sizes to be tested and their respective maximum jump limits.
	sizes := []mgu.Pair[int, int]{
		mgu.NewPair(Unigram, 0),               // Unigram: 0 jumps (always).
		mgu.NewPair(Bigram, MaxBigramJumps),   // Bigram: up to 4 jumps.
		mgu.NewPair(Trigram, MaxTrigramJumps), // Trigram: up to 2 jumps.
	}

	strB := strings.Builder{}
	strB.WriteString(csvHeader)

	// Main loop: iterates through each configured n-gram type (Unigram, Bigram, Trigram).
	for _, s := range sizes {
		size, maxJumps := s.Left, s.Right // Unpacks n-gram size and jump limit.

		// For each n-gram size, test all possible jump levels (from 0 to maxJumps).
		for jump := 0; jump <= maxJumps; jump++ {

			// --- OTIMIZAÇÃO: CRIA O AMBIENTE (DB + CACHE) UMA VEZ POR CONFIGURAÇÃO DE GRAM ---
			fmt.Printf(">>> Inicializando Ambiente para Size: %d, Jump: %d\n", size, jump)

			// Limpa o cache de memória antes de criar o novo cenário
			corpus.ResetCache()

			// Cria o banco de dados físico apenas UMA vez para este grupo de testes
			// Usamos o ID atual para nomear o arquivo, mas ele será reusado pelos próximos IDs
			dbName, dbConn := corpus.CreateDatabaseCaches(id, false, size, jump)

			// Loop interno para variações que NÃO exigem recriar o índice/banco
			for _, normalize := range []bool{false, true} {
				for _, parallel := range []bool{false, true} {

					// Execute TF-IDF reusing the DB
					strB.WriteString(BaseTest(id, dbConn, support.TdIdf, parallel, false, normalize, size, jump))
					id++

					// Execute BM25 reusing the DB
					strB.WriteString(BaseTest(id, dbConn, support.Bm25, parallel, false, normalize, size, jump))
					id++
				}
			}

			// --- CLEANUP: Desmonta o ambiente antes de mudar o tamanho do Gram/Jump ---

			// É crucial fechar a conexão SQL antes de tentar deletar o arquivo,
			// principalmente em Windows (Lock de arquivo).
			sqlDB, err := dbConn.DB()
			if err == nil {
				sqlDB.Close()
			}

			// Remove o arquivo físico do banco de dados criado para este grupo
			if err := os.Remove(dbName); err != nil {
				log.Printf("aviso: erro removendo arquivo de corpus %s: %v", dbName, err)
			}
			fmt.Printf("<<< Ambiente finalizado e limpo: %s\n", dbName)
		}
	}

	// --- Saving Results ---

	// Print an empty line for better console formatting.
	fmt.Println()
	fmt.Println(strB.String())

	err := os.WriteFile(ResultsOutput, []byte(strB.String()), 0644)
	if err != nil {
		log.Fatalf("error saving results.json: %v", err)
	}
}

// BaseTest executes a full benchmark and validation cycle using an EXISTING database connection.
// It no longer creates or deletes the database, only runs the algo logic.
func BaseTest(testId int64, db *gorm.DB, algo support.Algo, parallel, preIndexed, normalizeJumps bool, size, jumps int) string {

	// Nota: Não chamamos ResetCache() aqui para aproveitar o "aquecimento" do cache entre execuções parecidas
	// Nota: Não chamamos CreateDatabaseCaches() aqui, usamos o 'db' recebido

	legalInputs := "./../../misc/searchLegalInputs.json"

	// Passamos o DB já aberto
	res, err := corpus.ApplyLegalInputsDir(db, legalInputs, algo, preIndexed, normalizeJumps, parallel, size, jumps)
	if err != nil {
		log.Fatalf(err.Error())
	}

	// limpa o toString pra virar 1 linha
	clean := strings.ReplaceAll(res.String(), "\n", "")
	clean = strings.ReplaceAll(clean, "\t", "")

	csv := fmt.Sprintf(
		"%d,%s,%v,%v,%d,%d,%v,%s\n",
		testId, algo, preIndexed, normalizeJumps, size, jumps, parallel, clean,
	)

	return csv
}
