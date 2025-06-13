package utils

import (
	"regexp"
	"strings"
	"fmt"
	"sort"
)

const DEBUG = true

func TfIdfPreContext(input string) ([]string, error) {
	
	// Tira todas as minúsculas.
	input = strings.ToLower(input)
	
	// Regex para detectar números romanos e regex para detectar números decimais.
	// rx0 := regexp.MustCompile(`M{0,3}(CM|CD|D?C{0,3})(XC|XL|L?X{0,3})(IX|IV|V?I{0,3})`)
	rx1 := regexp.MustCompile(`\b((\d{1,3}(\.\d{3})*)|\d+)(,\d+)?\b`)
	
	// Substitui numerais por <n>
	// input = rx0.ReplaceAllString(input, "<n>")
	input = rx1.ReplaceAllString(input, "<n>")
	
	// Pega o array de caracteres especiais.
	replaceVec, err := LoadAccents()
	if err != nil {
		return nil, err
	}
	
	// Substitui todos os caracteres especiais.
	r := strings.NewReplacer(replaceVec...)
	result := r.Replace(input)
	
	// Separa a string em tokens por espaços.
	return strings.Split(result, " "), nil
}

func TFIDF(input string) ([]float64, error) {
	
	// Pega o conteúdo do arquivo
	content, err := "", error(nil)
	if strings.HasSuffix(strings.ToLower(input), ".pdf") {
		content, err = GetPDF(input)
	} else {
		content, err = GetTXT(input)
	}
	
	// Trata o conteúdo do arquivo
	tokens, err := TfIdfPreContext(content)
	if err != nil {
		return nil, err
	}
	
	// Calcula a frequência de cada token
	freqMap := make(map[string]uint)
	for _, token := range tokens {
		token = strings.TrimSpace(token)
		if token == "" {
			continue
		}
		freqMap[token]++
	}
	
	unique := make([]string, 0, len(freqMap))
	for tok := range freqMap {
		unique = append(unique, tok)
	}
	sort.Strings(unique)

	total := float64(len(tokens))
	ret := make([]float64, len(unique))
	for i, tok := range unique {
		tf := float64(freqMap[tok]) / total    // idf = 1 (single‑doc)
		ret[i] = tf
		if DEBUG {
			fmt.Printf("%s -> %.5f\n", tok, tf)
		}
	}
	return ret, nil
}