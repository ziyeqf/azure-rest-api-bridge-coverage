package runner

import (
	"strings"

	"bridge-coverage/jsonhelper"
	"github.com/go-openapi/jsonpointer"
)

type Runner struct {
	resources map[string]jsonhelper.ResourceJSON
	bridgeMap map[string]map[string]interface{}
	// map[resourceType]map[property]exist
	coverageResult map[string]map[string]bool
	scmCnt         map[string]int
	covCnt         map[string]int
}

func NwRunner(resources map[string]jsonhelper.ResourceJSON, bridgeMap map[string]map[string]interface{}) *Runner {
	return &Runner{
		resources:      resources,
		bridgeMap:      bridgeMap,
		coverageResult: make(map[string]map[string]bool),
		scmCnt:         make(map[string]int),
		covCnt:         make(map[string]int),
	}
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
	updateMap := func(ptrStr string) {
		if _, ok := r.coverageResult[resType]; !ok {
			r.coverageResult[resType] = make(map[string]bool)
		}

		if resourceMissed {
			r.coverageResult[resType][ptrStr] = false
			return
		}

		r.scmCnt[resType]++
		if _, ok := r.bridgeMap[resType][ptrStr]; ok {
			r.coverageResult[resType][ptrStr] = true
			r.covCnt[resType]++
		} else {
			r.coverageResult[resType][ptrStr] = false
		}
	}

	for n, sch := range schema {
		switch sch.Type {
		case jsonhelper.SchemaTypeList,
			jsonhelper.SchemaTypeSet,
			jsonhelper.SchemaTypeMap:
			// for nested properties
			switch t := sch.Elem.(type) {
			case string:
				jsonP, err := jsonpointer.New("/" + strings.Join(append(etkPrefix, n), "/") + "/0")
				if err != nil {
					return err
				}
				updateMap(jsonP.String())
			case jsonhelper.ResourceJSON:
				if err := r.HandleSchema(t.Schema, resType, append(etkPrefix, n, "0"), resourceMissed); err != nil {
					return err
				}
			}
		default:
			jsonP, err := jsonpointer.New("/" + strings.Join(append(etkPrefix, n), "/"))
			if err != nil {
				return err
			}
			updateMap(jsonP.String())
		}
	}

	return nil
}
