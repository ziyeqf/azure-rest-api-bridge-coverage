package jsonhelper

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
)

const (
	SchemaTypeSet  = "TypeSet"
	SchemaTypeList = "TypeList"
	SchemaTypeMap  = "TypeMap"
)

type SchemaJSON struct {
	Type string      `json:"type,omitempty"`
	Elem interface{} `json:"elem,omitempty"`
}

func (b *SchemaJSON) UnmarshalJSON(body []byte) error {
	var m map[string]interface{}
	err := json.Unmarshal(body, &m)
	if err != nil {
		return err
	}
	b.Type, _ = m["type"].(string)

	// todo: do we really need this? or just take it as a map to use is fine.
	if e, ok := m["elem"]; ok && e != nil {
		elem := e.(map[string]interface{})
		if schema, ok := elem["schema"]; ok {
			b.Elem = ResourceFromMap(schema.(map[string]interface{}))
		}
		if t, ok := elem["type"]; ok {
			b.Elem = t.(string)
		}
	}

	return nil
}

func ResourceFromMap(input map[string]interface{}) ResourceJSON {
	result := ResourceJSON{
		Schema: make(map[string]SchemaJSON, 0),
	}
	for k, v := range input {
		result.Schema[k] = SchemaFromMap(v.(map[string]interface{}))
	}
	return result
}

func SchemaFromMap(input map[string]interface{}) SchemaJSON {
	result := SchemaJSON{}
	if t, ok := input["type"]; ok {
		result.Type = t.(string)
	}

	return result
}

type ResourceJSON struct {
	Schema map[string]SchemaJSON `json:"schema"`
}

type ProviderSchemaJSON struct {
	ResourcesMap map[string]ResourceJSON `json:"resources,omitempty"`
}

type ProviderWrapper struct {
	ProviderSchema *ProviderSchemaJSON `json:"providerSchema,omitempty"`
}

func ParseSchema(path string) (*ProviderWrapper, error) {
	f, err := os.OpenFile(path, os.O_RDONLY, 0666)
	if err != nil {
		return nil, fmt.Errorf("open file: %v", err)
	}

	defer f.Close()

	jsonByte, err := io.ReadAll(f)
	if err != nil {
		return nil, fmt.Errorf("read file: %v", err)
	}

	var provider ProviderWrapper
	if err := json.Unmarshal(jsonByte, &provider); err != nil {
		return nil, fmt.Errorf("unmarshal json: %v", err)
	}

	return &provider, nil
}
