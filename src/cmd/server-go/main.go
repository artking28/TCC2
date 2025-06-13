package main

import (
	"log"

	"github.com/tcc2-davi-arthur/utils"
)

func main() {
	_, err := utils.TFIDF("../misc/file.pdf")
	if err != nil {
		log.Fatal(err)
	}
	
	log.Println("ok!")
}