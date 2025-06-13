package utils

import (
	"encoding/json"
	"os"
)

func LoadAccents() ([]string, error) {
	out, err := os.ReadFile("../misc/replaces.json")
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