package runner

import (
	"errors"
	"strings"

	"bridge-coverage/jsonhelper"
	"github.com/go-openapi/jsonpointer"
)

type Opts struct {
	Resources     map[string]jsonhelper.ResourceJSON
	BridgeMap     map[string]map[string]interface{}
	IgnoreSchemas []string
	MapIdentity   string
}
type Runner struct {
	resources     map[string]jsonhelper.ResourceJSON
	bridgeMap     map[string]map[string]interface{}
	ignoreSchemas []string
	mapIdentity   string
	// map[resourceType]map[property]exist
	coverageResult map[string]map[string]bool
	scmCnt         map[string]int
	covCnt         map[string]int
}

func NwRunner(opt Opts) (*Runner, error) {
	if opt.Resources == nil {
		return nil, errors.New("resources is nil")
	}
	if opt.BridgeMap == nil {
		return nil, errors.New("bridgeMap is nil")
	}

	return &Runner{
		resources:      opt.Resources,
		bridgeMap:      opt.BridgeMap,
		ignoreSchemas:  opt.IgnoreSchemas,
		mapIdentity:    opt.MapIdentity,
		coverageResult: make(map[string]map[string]bool),
		scmCnt:         make(map[string]int),
		covCnt:         make(map[string]int),
	}, nil
}

func (r Runner) Run() (details map[string]map[string]bool, schemaCnt map[string]int, coverageCnt map[string]int, err error) {
	for resType, res := range r.resources {
		resourceMissed := false
		if _, ok := r.bridgeMap[resType]; !ok {
			resourceMissed = true
		}

		if err := r.HandleSchema(res.Schema, resType, make([]string, 0), resourceMissed); err != nil {
			return nil, nil, nil, err
		}

	}
	return r.coverageResult, r.scmCnt, r.covCnt, nil
}

func (r Runner) HandleSchema(schema map[string]jsonhelper.SchemaJSON, resType string, etkPrefix []string, resourceMissed bool) error {
	updateMapFunc := func(ptrStr string) error {
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

		if resourceMissed {
			r.coverageResult[resType][ptrStr] = false
			return nil
		}

		r.scmCnt[resType]++
		if _, ok := r.bridgeMap[resType][ptrStr]; ok {
			r.coverageResult[resType][ptrStr] = true
			r.covCnt[resType]++
		} else {
			r.coverageResult[resType][ptrStr] = false
		}

		return nil
	}

	handleNestedFunc := func(elem interface{}, name string, childIdentity string) error {
		switch t := elem.(type) {
		case string:
			jsonP, err := jsonpointer.New("/" + strings.Join(append(etkPrefix, name), "/") + "/" + childIdentity)
			if err != nil {
				return err
			}
			if err := updateMapFunc(jsonP.String()); err != nil {
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
			if err := handleNestedFunc(sch.Elem, n, "0"); err != nil {
				return err
			}
		case jsonhelper.SchemaTypeMap:
			if err := handleNestedFunc(sch.Elem, n, r.mapIdentity); err != nil {
				return err
			}
		default:
			jsonP, err := jsonpointer.New("/" + strings.Join(append(etkPrefix, n), "/"))
			if err != nil {
				return err
			}
			if err := updateMapFunc(jsonP.String()); err != nil {
				return err
			}
		}
	}

	return nil
}
