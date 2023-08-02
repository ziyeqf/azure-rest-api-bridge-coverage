package jsonhelper

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
)

type PropertyCoverage struct {
	Addr       string `json:"addr"`
	LinkGithub string `json:"link_github"`
	LinkLocal  string `json:"link_local"`
	Ref        string `json:"ref"`
}

// return map[reseourceType]map[appAddr][]PropertyCoverage
func ParseCoverageFile(path string) (map[string]map[string][]PropertyCoverage, error) {
	f, err := os.OpenFile(path, os.O_RDONLY, 0666)
	if err != nil {
		return nil, fmt.Errorf("open file: %v", err)
	}

	defer f.Close()

	jsonByte, err := io.ReadAll(f)
	if err != nil {
		return nil, fmt.Errorf("read file: %v", err)
	}

	var coverageMap map[string]map[string][]PropertyCoverage
	//var foo map[string]map[string]interface{}
	if err := json.Unmarshal(jsonByte, &coverageMap); err != nil {
		return nil, fmt.Errorf("unmarshal json: %v", err)
	}

	return coverageMap, nil
}
