package runner

import (
	"errors"
	"strings"

	"github.com/go-openapi/jsonpointer"
	"terraform-azurerm-provider-coverage/jsonhelper"
)

type Opts struct {
	Resources                map[string]jsonhelper.ResourceJSON
	CoverageMap              map[string]map[string]interface{}
	IgnoreSchemas            []string
	PrefixMatch              bool
	IgnoreUncoveredResources bool
}
type Runner struct {
	resources                map[string]jsonhelper.ResourceJSON
	coverageMap              map[string]map[string]interface{}
	ignoreSchemas            []string
	prefixMatch              bool
	ignoreUncoveredResources bool
	// map[resourceType]map[property]exist
	coverageResult map[string]map[string]bool
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

	return &Runner{
		resources:                opt.Resources,
		coverageMap:              opt.CoverageMap,
		ignoreSchemas:            opt.IgnoreSchemas,
		prefixMatch:              opt.PrefixMatch,
		ignoreUncoveredResources: opt.IgnoreUncoveredResources,
		coverageResult:           make(map[string]map[string]bool),
		scmCnt:                   make(map[string]int),
		covCnt:                   make(map[string]int),
	}, nil
}

func (r Runner) Run() (details map[string]map[string]bool, schemaCnt map[string]int, coverageCnt map[string]int, err error) {
	for resType, res := range r.resources {
		resourceMissed := false
		if coverage, ok := r.coverageMap[resType]; !ok {
			resourceMissed = true
		} else {
			resourceMissed = len(coverage) == 0
		}

		if r.ignoreUncoveredResources && resourceMissed {
			continue
		}

		if err := r.HandleSchema(res.Schema, resType, make([]string, 0), resourceMissed); err != nil {
			return nil, nil, nil, err
		}

	}
	return r.coverageResult, r.scmCnt, r.covCnt, nil
}

func (r Runner) HandleSchema(schema map[string]jsonhelper.SchemaJSON, resType string, etkPrefix []string, resourceMissed bool) error {
	// prefixMatch is special for map, as TypeString in map might has key name in coverage
	// e.g.:
	// coverage: "/example/key" "/example/abc" etc
	// schema:
	// "example":{
	//    "type": "TypeMap",
	//    "optional": true,
	//    "forceNew": true,
	//    "elem": {
	//        "type": "TypeString"
	//    }
	//}
	// then we will need to use the prefix "/example" to match
	updateMapFunc := func(ptrStr string, prefixMatch bool) error {
		if len(r.ignoreSchemas) > 0 {
			for _, ignoreSchema := range r.ignoreSchemas {
				ignorePtr, err := jsonpointer.New("/" + ignoreSchema)
				if err != nil {
					return err
				}
				if ptrStr == ignorePtr.String() {
					return nil
				}
			}
		}

		if _, ok := r.coverageResult[resType]; !ok {
			r.coverageResult[resType] = make(map[string]bool)
		}

		r.scmCnt[resType]++

		if resourceMissed {
			r.coverageResult[resType][ptrStr] = false
			return nil
		}

		if prefixMatch {
			r.coverageResult[resType][ptrStr] = false
			for k, _ := range r.coverageMap[resType] {
				if strings.HasPrefix(k, ptrStr) {
					r.coverageResult[resType][ptrStr] = true
					r.covCnt[resType]++
					break
				}
			}
			return nil
		}

		if _, ok := r.coverageMap[resType][ptrStr]; ok {
			r.coverageResult[resType][ptrStr] = true
			r.covCnt[resType]++
		} else {
			r.coverageResult[resType][ptrStr] = false
		}
		return nil
	}

	handleNestedFunc := func(elem interface{}, name string, prefixMatch bool) error {
		ptr := "/" + strings.Join(append(etkPrefix, name), "/")
		if !prefixMatch {
			ptr += "/0"
		}
		switch t := elem.(type) {
		case string:
			jsonP, err := jsonpointer.New(ptr)
			if err != nil {
				return err
			}
			if err := updateMapFunc(jsonP.String(), prefixMatch); err != nil {
				return err
			}
		case jsonhelper.ResourceJSON:
			if err := r.HandleSchema(t.Schema, resType, append(etkPrefix, name, "0"), resourceMissed); err != nil {
				return err
			}
		}
		return nil
	}

	for n, sch := range schema {
		switch sch.Type {
		case jsonhelper.SchemaTypeList,
			jsonhelper.SchemaTypeSet:
			if err := handleNestedFunc(sch.Elem, n, false); err != nil {
				return err
			}
		case jsonhelper.SchemaTypeMap:
			if err := handleNestedFunc(sch.Elem, n, true); err != nil {
				return err
			}
		default:
			jsonP, err := jsonpointer.New("/" + strings.Join(append(etkPrefix, n), "/"))
			if err != nil {
				return err
			}
			if err := updateMapFunc(jsonP.String(), false); err != nil {
				return err
			}
		}
	}

	return nil
}
