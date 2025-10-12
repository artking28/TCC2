package utils

import (
	"encoding/json"
	"os"
	"regexp"
	"strings"
	"unicode"

	"github.com/bbalet/stopwords"
)

func LoadAccents() ([]string, error) {
	out, err := os.ReadFile("misc/replaces.json")
	if err != nil {
		return nil, err
	}

	var m map[string][]string
	err = json.Unmarshal(out, &m)
	if err != nil {
		return nil, err
	}

	var ret []string
	for key, all := range m {
		for _, token := range all {
			ret = append(ret, token, key)
		}
	}

	return ret, nil
}

func CleanText(input string) (string, error) {

	// Tira todas as minúsculas.
	input = strings.ToLower(input)

	// Normalizações
	input = strings.ReplaceAll(input, "–", " ")
	input = strings.ReplaceAll(input, "-", " ")
	input = strings.ReplaceAll(input, "_", " ")

	// Remove pontuações
	input = strings.Map(func(r rune) rune {
		if unicode.IsPunct(r) || unicode.IsSymbol(r) {
			return -1
		}
		return r
	}, input)

	// Regex para detectar números romanos e regex para detectar números decimais.
	rx0 := regexp.MustCompile(`^M{0,3}(CM|CD|D?C{0,3})(XC|XL|L?X{0,3})(IX|IV|V?I{0,3})$`)
	rx1 := regexp.MustCompile(`\b((\d{1,3}(\.\d{3})*)|\d+)(,\d+)?\b`)
	rx2 := regexp.MustCompile(`\b((\d{1,3}(,\d{3})*)|\d+)(\.\d+)?\b`)
	rx3 := regexp.MustCompile(`<número>(( |\.)?<número>)*`)
	rx4 := regexp.MustCompile(`\b(\d{1,2}[-/]\d{1,2}[-/]\d{2,4}|\d{4}[-/]\d{1,2}[-/]\d{1,2})\b`)

	// Substitui numerais por <n>
	input = rx0.ReplaceAllString(input, "número")
	input = rx1.ReplaceAllString(input, "número")
	input = rx2.ReplaceAllString(input, "número")
	input = rx3.ReplaceAllString(input, "código")
	input = rx4.ReplaceAllString(input, "data")

	// Pega o array de caracteres especiais.
	replaceVec, err := LoadAccents()
	if err != nil {
		return "", err
	}

	// Substitui todos os caracteres especiais.
	r := strings.NewReplacer(replaceVec...)
	input = r.Replace(input)

	input = stopwords.CleanString(input, "pt", false)
	return input, nil
}
