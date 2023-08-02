package jsonhelper

import (
	"fmt"

	"github.com/go-openapi/jsonpointer"
)

type ResourceOutput struct {
	Name            string     `json:"name"`
	TotalCnt        int        `json:"total_cnt"`
	CoveredCnt      int        `json:"covered_cnt"`
	UncoveredCnt    int        `json:"uncovered_cnt"`
	CoveredPercent  string     `json:"covered_percent"`
	CoveredFields   SchemaNode `json:"covered_fields"`
	UncoveredFields SchemaNode `json:"uncovered_fields"`
}

type SchemaNode struct {
	RootChildren map[string]FieldOutput `json:"root_children,omitempty"`
	Children     map[string]SchemaNode  `json:"children,omitempty"`
}

type FieldOutput struct {
	GithubUrl string `json:"github_url"`
}

func (root SchemaNode) fillFields(tks []string, detail *PropertyCoverage) SchemaNode {
	if len(tks) == 1 {
		if root.RootChildren == nil {
			root.RootChildren = make(map[string]FieldOutput, 0)
		}
		if detail != nil {
			root.RootChildren[tks[0]] = FieldOutput{
				GithubUrl: detail.LinkGithub,
			}
		} else {
			root.RootChildren[tks[0]] = FieldOutput{}
		}

		//root.RootChildren = append(root.RootChildren, tks[0])
		return root
	}

	if root.Children == nil {
		root.Children = make(map[string]SchemaNode)
	}

	if _, ok := root.Children[tks[0]]; !ok {
		root.Children[tks[0]] = SchemaNode{}
	}

	root.Children[tks[0]] = root.Children[tks[0]].fillFields(tks[1:], detail)
	return root
}

func GenResourceOutput(name string, fieldsCoverageMap map[string]*PropertyCoverage) (ResourceOutput, error) {
	output := ResourceOutput{
		Name: name,
		CoveredFields: SchemaNode{
			RootChildren: make(map[string]FieldOutput, 0),
		},
		UncoveredFields: SchemaNode{
			RootChildren: make(map[string]FieldOutput, 0),
		},
	}

	for name, detail := range fieldsCoverageMap {
		jptr, err := jsonpointer.New(name)
		if err != nil {
			return output, err
		}
		tks := make([]string, 0)
		// TODO: remove "0" to make it fit the portal
		for _, tk := range jptr.DecodedTokens() {
			tks = append(tks, tk)
		}

		if len(tks) == 1 {
			output.TotalCnt++
			//tkName := tks[0]
			if detail != nil {
				output.CoveredCnt++
				output.CoveredFields.RootChildren[tks[0]] = FieldOutput{
					GithubUrl: detail.LinkGithub,
				}
				//output.CoveredFields.RootChildren = append(output.CoveredFields.RootChildren, FieldOutput{GithubUrl: detail})
			} else {
				output.UncoveredCnt++
				output.UncoveredFields.RootChildren[tks[0]] = FieldOutput{}
				//output.UncoveredFields.RootChildren = append(output.UncoveredFields.RootChildren, tkName)
			}
		} else {
			if detail != nil {
				output.CoveredFields = output.CoveredFields.fillFields(tks, detail)
			} else {
				output.UncoveredFields = output.UncoveredFields.fillFields(tks, detail)
			}
		}
	}

	output.CoveredPercent = fmt.Sprintf("%.2f%%", float32(output.CoveredCnt)/float32(output.TotalCnt)*100)
	return output, nil
}

type PortalDiagnosticOutput struct {
	TotalCoverPercent string                `json:"total_cover_percent"`
	TotalFields       int                   `json:"total_fields"`
	TotalCovered      int                   `json:"total_covered"`
	TotalResources    int                   `json:"total_resources"`
	IssueResource     []PortalIssueResource `json:"issue_resource"`
}

type PortalIssueResource struct {
	Name         string `json:"name"`
	StaticsCount int    `json:"statics_count"`
	CoveredCount int    `json:"covered_count"`
}

func GenPortalDiagnosticOutput(covCnt, scmCnt map[string]int, ignoreUncoveredResources *bool, coverageMap map[string]map[string][]PropertyCoverage) PortalDiagnosticOutput {
	totalScm := 0
	totalCov := 0

	for _, v := range covCnt {
		totalCov += v
	}

	for _, v := range scmCnt {
		totalScm += v
	}

	resultCnt := scmCnt
	if *ignoreUncoveredResources {
		resultCnt = covCnt
	}

	issueRes := make([]PortalIssueResource, 0)
	for k := range resultCnt {
		if covCnt[k] != len(coverageMap[k]) {
			issueRes = append(issueRes, PortalIssueResource{
				Name:         k,
				StaticsCount: covCnt[k],
				CoveredCount: len(coverageMap[k]),
			})
		}
	}

	return PortalDiagnosticOutput{
		IssueResource:     issueRes,
		TotalResources:    len(scmCnt),
		TotalCovered:      totalCov,
		TotalFields:       totalScm,
		TotalCoverPercent: fmt.Sprintf("%.2f%%", float32(totalCov)/float32(totalScm)*100),
	}
}
