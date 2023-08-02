package runner

import (
	"errors"
	"fmt"
	"sort"

	"github.com/go-openapi/jsonpointer"
	"terraform-azurerm-provider-coverage/jsonhelper"
	"terraform-azurerm-provider-coverage/jsontree"
)

type Opts struct {
	Resources                map[string]jsonhelper.ResourceJSON
	CoverageMap              map[string]map[string][]jsonhelper.PropertyCoverage
	IgnoreSchemas            []string
	IgnoreUncoveredResources bool
}

type Runner struct {
	resources                map[string]jsonhelper.ResourceJSON
	coverageMap              map[string]map[string][]jsonhelper.PropertyCoverage
	ignoreSchemas            []string
	ignoreUncoveredResources bool
	parsedCoverageTree       map[string]*jsontree.Node
	// map[resourceType]map[property]coverage_detail
	coverageResult map[string]map[string]*jsonhelper.PropertyCoverage
	scmCnt         map[string]int
	covCnt         map[string]int
}

func NwRunner(opt Opts) (*Runner, error) {
	if opt.Resources == nil {
		return nil, errors.New("resources is nil")
	}
	if opt.CoverageMap == nil {
		return nil, errors.New("coverageMap is nil")
	}

	parsedCoverageTree := make(map[string]*jsontree.Node)
	for n, res := range opt.CoverageMap {
		rootNode := jsontree.NewNode("/")
		parsedCoverageTree[n] = &rootNode
		for prop := range res {
			ptr, err := jsonpointer.New(prop)
			if err != nil {
				return nil, err
			}
			rootNode, err = jsontree.ParseJsonPtr(&rootNode, ptr)
			if err != nil {
				return nil, err
			}
		}
	}

	return &Runner{
		resources:                opt.Resources,
		coverageMap:              opt.CoverageMap,
		ignoreSchemas:            opt.IgnoreSchemas,
		ignoreUncoveredResources: opt.IgnoreUncoveredResources,
		coverageResult:           make(map[string]map[string]*jsonhelper.PropertyCoverage),
		scmCnt:                   make(map[string]int),
		covCnt:                   make(map[string]int),
		parsedCoverageTree:       parsedCoverageTree,
	}, nil
}

// details map[resourceType]map[Property]coverageDetail
// for non-exist property, reference is nil
func (r Runner) Run() (details map[string]map[string]*jsonhelper.PropertyCoverage, schemaCnt map[string]int, coverageCnt map[string]int, err error) {
	for resType, res := range r.resources {
		resourceMissed := false
		if resource, ok := r.coverageMap[resType]; !ok {
			resourceMissed = true
		} else {
			resourceMissed = len(resource) == 0
		}

		if r.ignoreUncoveredResources && resourceMissed {
			continue
		}

		resCtx := ResourceContext{
			Name:               resType,
			Schema:             res.Schema,
			TokenPrefix:        make([]string, 0),
			DisplayTokenPrefix: make([]string, 0),
		}

		if err := r.HandleSchema(resCtx); err != nil {
			return nil, nil, nil, err
		}

	}
	return r.coverageResult, r.scmCnt, r.covCnt, nil
}

func (r Runner) HandleNestedSchema(resCtx ResourceContext) func(elem interface{}, name string) error {
	return func(elem interface{}, name string) error {
		ptr, err := resCtx.JsonPtr(name)
		if err != nil {
			return err
		}

		possibleNames, _ := r.GetAllChildrenNames(resCtx.Name, ptr)
		switch t := elem.(type) {
		case string:
			if len(possibleNames) == 0 {
				possibleNames = append(possibleNames, "/0")
				possibleNames = append(possibleNames, "")
			}
			for _, n := range possibleNames {
				ptr, err := resCtx.JsonPtr(name + "/" + n)
				if err != nil {
					return err
				}
				displayPtr, err := resCtx.DisplayJsonPtr(name)
				if err != nil {
					return err
				}
				if err := resCtx.UpdateMapFunc(ptr, displayPtr); err != nil {
					return err
				}
			}
		case jsonhelper.ResourceJSON:
			if len(possibleNames) == 0 {
				possibleNames = append(possibleNames, "0")
			}
			for _, n := range possibleNames {
				if err := r.HandleSchema(resCtx.update(t.Schema, []string{name, n}, []string{name, possibleNames[0]})); err != nil {
					return err
				}
			}
		}
		return nil
	}
}

func (r Runner) HandleSchema(resCtx ResourceContext) error {
	resCtx.UpdateMapFunc = r.UpdateCoverageResult(resCtx.Name)
	handleNestedFunc := r.HandleNestedSchema(resCtx)

	for n, sch := range resCtx.Schema {
		switch sch.Type {
		case jsonhelper.SchemaTypeList,
			jsonhelper.SchemaTypeSet,
			jsonhelper.SchemaTypeMap:
			if err := handleNestedFunc(sch.Elem, n); err != nil {
				return err
			}
		default:
			ptr, err := resCtx.JsonPtr(n)
			if err != nil {
				return err
			}
			displayPtr, err := resCtx.DisplayJsonPtr(n)
			if err != nil {
				return err
			}
			if err := resCtx.UpdateMapFunc(ptr, displayPtr); err != nil {
				return err
			}
		}
	}

	return nil
}

func (r Runner) UpdateCoverageResult(resType string) func(ptrStr string, realPtrStr string) error {
	return func(ptrStr string, displayPtrStr string) error {
		if len(r.ignoreSchemas) > 0 {
			for _, ignoreSchema := range r.ignoreSchemas {
				ignorePtr, err := jsonpointer.New("/" + ignoreSchema)
				if err != nil {
					return err
				}
				if ptrStr == ignorePtr.String() {
					return nil
				}
				if displayPtrStr == ignorePtr.String() {
					return nil
				}
			}
		}

		if _, ok := r.coverageResult[resType]; !ok {
			r.coverageResult[resType] = make(map[string]*jsonhelper.PropertyCoverage)
		}

		detail, ok := r.coverageMap[resType][ptrStr]
		r.UpdatePropExist(resType, displayPtrStr, ok, detail)

		return nil
	}

}

// never use `false` to override `true` on the result map.
func (r Runner) UpdatePropExist(resType string, propPtr string, exist bool, detail []jsonhelper.PropertyCoverage) {
	if e, ok := r.coverageResult[resType][propPtr]; ok && e != nil {
		return
	}

	if exist {
		r.coverageResult[resType][propPtr] = &detail[0]
		r.covCnt[resType]++
	} else {
		r.coverageResult[resType][propPtr] = nil
	}
	r.scmCnt[resType]++
}

func (r Runner) GetAllChildrenNames(resType, ptrStr string) ([]string, error) {
	root, ok := r.parsedCoverageTree[resType]
	if !ok {
		return nil, fmt.Errorf("resource type %v not found on coverage tree", resType)
	}

	ptr, err := jsonpointer.New(ptrStr)
	if err != nil {
		return nil, err
	}

	cur := root
	for _, tk := range ptr.DecodedTokens() {
		if _, ok := cur.Children[tk]; !ok {
			return nil, fmt.Errorf("no node match %v", tk)
		}

		n, _ := cur.Children[tk]
		cur = &n
	}

	result := make([]string, 0)
	for k := range cur.Children {
		result = append(result, k)
	}

	// do sort to keep result always same
	sort.Strings(result)
	return result, nil
}
