package utils

import (
	"os"
	"os/exec"
	"strings"
)

func GetPDF(file string) ([]string, error) {
	out, err := exec.Command("pdftotext", file, "-").Output()
	if err != nil {
		return nil, err
	}
	
	return PreContext(string(out))
}

func GetTXT(file string) ([]string, error) {
	out, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}
	
	return PreContext(string(out))
}

func PreProcessTXT(file string, output string) error {
	out, err := os.ReadFile(file)
	if err != nil {
		return err
	}
	
	result, err := PreContext(string(out))
	if err != nil {
		return err
	}
	
	return os.WriteFile(output, []byte(strings.Join(result, " ")), 0744)
}