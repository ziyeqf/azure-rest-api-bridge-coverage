package jsonhelper

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
)

// return map[reseourceType]map[appAddr]PropertyMap
func ParseCoverageFile(path string) (map[string]map[string]interface{}, error) {
	f, err := os.OpenFile(path, os.O_RDONLY, 0666)
	if err != nil {
		return nil, fmt.Errorf("open file: %v", err)
	}

	defer f.Close()

	jsonByte, err := io.ReadAll(f)
	if err != nil {
		return nil, fmt.Errorf("read file: %v", err)
	}

	var bridgeMap map[string]map[string]interface{}
	if err := json.Unmarshal(jsonByte, &bridgeMap); err != nil {
		return nil, fmt.Errorf("unmarshal json: %v", err)
	}

	return bridgeMap, nil
}
