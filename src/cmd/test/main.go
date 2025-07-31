package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/tcc2-davi-arthur/utils"
	"golang.org/x/term"
)

var wordMap = make(map[string]bool)

func beforeDestroy(fd int,  oldState *term.State) {
	fmt.Println("\n[beforeDestroy]")
	term.Restore(fd, oldState)
	
	data, _ := json.MarshalIndent(wordMap, "", "   ")
	os.WriteFile("misc/cache.json", data, 0744)
	
	os.Exit(0)
}

func main() {
	if !term.IsTerminal(int(os.Stdin.Fd())) {
		fmt.Println("stdin não é terminal")
		return
	}

	txtName := "misc/vidasSecas.txt"
	words, err := utils.GetTXT(txtName)
	if err != nil {
		panic(err)
	}

	fd := int(os.Stdin.Fd())
	oldState, err := term.MakeRaw(fd)
	if err != nil {
		fmt.Println("Erro ativando modo raw:", err)
		return
	}

	// handle sinais — executa cleanup direto aqui!
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP, syscall.SIGQUIT)
	go func() {
		<-sigs
		beforeDestroy(fd, oldState)
	}()

	// defer para finalização normal
	defer func() {
		beforeDestroy(fd, oldState)
	}()
	
	pos := 0
	k:=0
	for pos < len(words) {
		word := words[pos]

		if _, ok := wordMap[word]; ok {
			pos++
			continue
		}

		fmt.Printf("Palavra (%d | %d) ['%s']?", pos, k, word)

		var buf [1]byte
		_, err := syscall.Read(syscall.Stdin, buf[:])
		if err != nil {
			fmt.Println("Erro lendo tecla:", err)
			break
		}

		switch buf[0] {
		case 3:
			beforeDestroy(fd, oldState)
		case '\r', '\n':
			wordMap[word] = true
			fmt.Print("\033[1;32m Accepted\033[0m \r\n") // verde e negrito
			pos++
			k++
		case 'Z', 'z':
			if pos > 0 {
				pos--
				k--
				delete(wordMap, words[pos])
				fmt.Print("\033[1;34m Going back\033[0m \r\n")
			}
		default:
			wordMap[word] = false
			fmt.Print("\033[1;31m Rejected\033[0m \r\n")
			pos++
			k++
		}
	}
}
