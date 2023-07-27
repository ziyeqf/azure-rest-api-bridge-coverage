package jsonhelper

import (
	"fmt"

	"github.com/go-openapi/jsonpointer"
)

type ResourceOutput struct {
	TotalCnt        int        `json:"total_cnt"`
	CoveredCnt      int        `json:"covered_cnt"`
	UncoveredCnt    int        `json:"uncovered_cnt"`
	CoveredPercent  string     `json:"covered_percent"`
	CoveredFields   SchemaNode `json:"covered_fields"`
	UncoveredFields SchemaNode `json:"uncovered_fields"`
}

type SchemaNode struct {
	RootChildren []string              `json:"root_children,omitempty"`
	Children     map[string]SchemaNode `json:"children,omitempty"`
}

func (root SchemaNode) fillFields(tks []string) SchemaNode {
	if len(tks) == 1 {
		if root.RootChildren == nil {
			root.RootChildren = make([]string, 0)
		}
		root.RootChildren = append(root.RootChildren, tks[0])
		return root
	}

	if root.Children == nil {
		root.Children = make(map[string]SchemaNode)
	}

	if _, ok := root.Children[tks[0]]; !ok {
		root.Children[tks[0]] = SchemaNode{}
	}

	root.Children[tks[0]] = root.Children[tks[0]].fillFields(tks[1:])
	return root
}

func GenResourceOutput(fieldsCoverageMap map[string]bool) (ResourceOutput, error) {
	output := ResourceOutput{
		CoveredFields:   SchemaNode{},
		UncoveredFields: SchemaNode{},
	}

	for name, exist := range fieldsCoverageMap {
		jptr, err := jsonpointer.New(name)
		if err != nil {
			return output, err
		}
		if len(jptr.DecodedTokens()) == 1 {
			output.TotalCnt++
			tkName := jptr.DecodedTokens()[0]
			if exist {
				output.CoveredCnt++
				output.CoveredFields.RootChildren = append(output.CoveredFields.RootChildren, tkName)
			} else {
				output.UncoveredCnt++
				output.UncoveredFields.RootChildren = append(output.UncoveredFields.RootChildren, tkName)
			}
		} else {
			if exist {
				output.CoveredFields = output.CoveredFields.fillFields(jptr.DecodedTokens())
			} else {
				output.UncoveredFields = output.UncoveredFields.fillFields(jptr.DecodedTokens())
			}
		}
	}

	output.CoveredPercent = fmt.Sprintf("%.2f%%", float32(output.CoveredCnt)/float32(output.TotalCnt)*100)
	return output, nil
}
