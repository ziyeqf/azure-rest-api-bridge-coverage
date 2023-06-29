package runner

import (
	"strings"

	"bridge-coverage/jsonhelper"
	"github.com/go-openapi/jsonpointer"
)

type Runner struct {
	resources map[string]jsonhelper.ResourceJSON
	bridgeMap map[string]map[string]jsonhelper.PropertyMap
	// map[resourceType]map[property]exist
	coverageResult map[string]map[string]bool
}

func NwRunner(resources map[string]jsonhelper.ResourceJSON, bridgeMap map[string]map[string]jsonhelper.PropertyMap) *Runner {
	return &Runner{
		resources:      resources,
		bridgeMap:      bridgeMap,
		coverageResult: make(map[string]map[string]bool),
	}
}

func (r Runner) Run() (map[string]map[string]bool, error) {
	for resType, res := range r.resources {
		resourceMissed := false
		if _, ok := r.bridgeMap[resType]; !ok {
			resourceMissed = true
		}

		if err := r.HandleSchema(res.Schema, resType, make([]string, 0), resourceMissed); err != nil {
			return nil, err
		}

	}
	return r.coverageResult, nil
}

func (r Runner) HandleSchema(schema map[string]jsonhelper.SchemaJSON, resType string, etkPrefix []string, resourceMissed bool) error {
	updateMap := func(ptrStr string) {
		if _, ok := r.coverageResult[resType]; !ok {
			r.coverageResult[resType] = make(map[string]bool)
		}

		if resourceMissed {
			r.coverageResult[resType][ptrStr] = false
		}

		if _, ok := r.bridgeMap[resType][ptrStr]; ok {
			r.coverageResult[resType][ptrStr] = true
		} else {
			r.coverageResult[resType][ptrStr] = false
		}
	}

	for n, sch := range schema {
		switch sch.Type {
		case jsonhelper.SchemaTypeList:
		case jsonhelper.SchemaTypeSet:
		case jsonhelper.SchemaTypeMap:
			// for nested properties
			switch t := sch.Elem.(type) {
			case string:
				jsonP, err := jsonpointer.New("/" + strings.Join(append(etkPrefix, n), "/") + "/0")
				if err != nil {
					return err
				}
				updateMap(jsonP.String())
			case jsonhelper.ResourceJSON:
				if err := r.HandleSchema(t.Schema, resType, append(etkPrefix, n), resourceMissed); err != nil {
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
