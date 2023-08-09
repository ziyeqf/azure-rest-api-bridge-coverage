package runner

import (
	"strings"

	"github.com/go-openapi/jsonpointer"
	"github.com/ziyeqf/terraform-azurerm-provider-coverage/jsonhelper"
)

type ResourceContext struct {
	Name        string // resource type
	Schema      map[string]jsonhelper.SchemaJSON
	TokenPrefix []string
	// display token prefix is less than token prefix
	// because it might generate duplicate ptr when meet map and array.
	DisplayTokenPrefix []string
	UpdateMapFunc      func(ptrStr string, realPtrStr string) error
}

func (res ResourceContext) update(schema map[string]jsonhelper.SchemaJSON, newToken []string, newDisplayToken []string) ResourceContext {
	res.Schema = schema
	res.TokenPrefix = append(res.TokenPrefix, newToken...)
	res.DisplayTokenPrefix = append(res.DisplayTokenPrefix, newDisplayToken...)
	return res
}

func (res ResourceContext) JsonPtr(name string) (string, error) {
	jsonP, err := jsonpointer.New("/" + strings.Join(append(res.TokenPrefix, name), "/"))
	if err != nil {
		return "", err
	}
	return jsonP.String(), nil
}

func (res ResourceContext) DisplayJsonPtr(name string) (string, error) {
	jsonP, err := jsonpointer.New("/" + strings.Join(append(res.DisplayTokenPrefix, name), "/"))
	if err != nil {
		return "", err
	}
	return jsonP.String(), nil
}
