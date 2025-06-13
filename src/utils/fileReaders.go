package utils

import (
	"os"
	"os/exec"
)


func GetPDF(file string) (string, error) {
	out, err := exec.Command("pdftotext", file, "-").Output()
	if err != nil {
		return "", err
	}
	
	return string(out), nil
}

func GetTXT(file string) (string, error) {
	out, err := os.ReadFile(file)
	if err != nil {
		return "", err
	}
	
	return string(out), nil
}