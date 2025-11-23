package utils

import (
	"encoding/json"
	"errors"
	"io"
	"os"
	"regexp"
	"runtime"
	"strings"
	"time"
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

func Stopwatch(fn func()) time.Duration {
	start := time.Now()
	fn()
	return time.Since(start)
}

func MeasureMemory(fn func()) (uint64, uint64, uint64) {
	var minn, maxx, sum uint64
	var count uint64

	ticker := time.NewTicker(30 * time.Millisecond)
	done := make(chan struct{})

	go func() {
		for range ticker.C {
			var m runtime.MemStats
			runtime.ReadMemStats(&m)
			alloc := m.Alloc

			if minn == 0 || alloc < minn {
				minn = alloc
			}
			if alloc > maxx {
				maxx = alloc
			}
			sum += alloc
			count++
		}
	}()

	fn() // executa a função-alvo

	close(done)
	ticker.Stop()

	avg := uint64(0)
	if count > 0 {
		avg = sum / count
	}

	return minn, maxx, avg
}

func CleanText(input string) (string, error) {

	// Tira todas as minúsculas.
	input = strings.ToLower(input)

	// Regex para detectar números romanos e regex para detectar números decimais.
	rx0 := regexp.MustCompile(`^M{0,3}(CM|CD|D?C{0,3})(XC|XL|L?X{0,3})(IX|IV|V?I{0,3})$`)
	rx1 := regexp.MustCompile(`((\d{1,3}(\.\d{3})*)|\d+)(,\d+)?\b`)
	rx2 := regexp.MustCompile(`((\d{1,3}(,\d{3})*)|\d+)(\.\d+)?\b`)
	rx4 := regexp.MustCompile(`(\d{1,2}[-/]\d{1,2}[-/]\d{2,4}|\d{4}[-/]\d{1,2}[-/]\d{1,2})\b`)
	rx5 := regexp.MustCompile(`https?://([^\s]|\n)+`)

	// Substitui numerais por <n>
	input = rx5.ReplaceAllString(input, "")
	input = rx0.ReplaceAllString(input, "")
	input = rx1.ReplaceAllString(input, "")
	input = rx2.ReplaceAllString(input, "")
	input = rx4.ReplaceAllString(input, "")

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

// DuplicateFile copia um arquivo de src para dst.
// Retorna erro se algo falhar.
func DuplicateFile(src, dst string) (err error) {
	source, err := os.Open(src)
	if err != nil {
		return err
	}
	defer func() { err = errors.Join(err, source.Close()) }()

	dest, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer func() { err = errors.Join(err, dest.Close()) }()

	if _, err = io.Copy(dest, source); err != nil {
		return err
	}

	info, err := os.Stat(src)
	if err != nil {
		return err
	}

	return os.Chmod(dst, info.Mode())
}
