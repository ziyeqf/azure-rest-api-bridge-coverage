package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"sort"
	"strings"

	"terraform-azurerm-provider-coverage/jsonhelper"
	"terraform-azurerm-provider-coverage/runner"
)

func main() {
	coverageFile := flag.String("input", "", "the input file of schema")
	schemaFile := flag.String("schema", "", "the schema dump of azurerm provider")
	ignoreSchemas := flag.String("ignore-schema", "", "the schema to ignore of azurerm provider")
	ignoreUncoveredResources := flag.Bool("ignore-uncovered-resources", false, "ignore uncovered resources")
	diagnosticsOutput := flag.Bool("diagnostics-output", false, "output diagnostics information")
	portalOutput := flag.Bool("portal-output", false, "output to fit portal format")
	flag.Parse()

	coverageMap, err := jsonhelper.ParseCoverageFile(*coverageFile)
	if err != nil {
		exitOnError(err)
	}

	schema, err := jsonhelper.ParseSchema(*schemaFile)
	if err != nil {
		exitOnError(err)
	}

	ignoreSchemaList := make([]string, 0)
	if ignoreSchemas != nil && *ignoreSchemas != "" {
		ignoreSchemaList = append(ignoreSchemaList, strings.Split(*ignoreSchemas, ",")...)
	}

	r, err := runner.NwRunner(runner.Opts{
		Resources:                schema.ProviderSchema.ResourcesMap,
		CoverageMap:              coverageMap,
		IgnoreSchemas:            ignoreSchemaList,
		IgnoreUncoveredResources: *ignoreUncoveredResources,
	})
	if err != nil {
		exitOnError(err)
	}

	detail, scmCnt, covCnt, err := r.Run()
	if err != nil {
		exitOnError(err)
	}

	var output interface{}
	if !*portalOutput {
		o := make(map[string]map[string][]string)
		for k, v := range detail {
			o[k] = make(map[string][]string)
			o[k]["covered_properties"] = make([]string, 0)
			o[k]["uncovered_properties"] = make([]string, 0)
			for name, exist := range v {
				if exist {
					o[k]["covered_properties"] = append(o[k]["covered_properties"], name)
				} else {
					o[k]["uncovered_properties"] = append(o[k]["uncovered_properties"], name)
				}
			}
		}
		output = o
		if *diagnosticsOutput {
			diagOutput(covCnt, scmCnt, ignoreUncoveredResources, coverageMap)
		}
	} else {
		resources := make([]jsonhelper.ResourceOutput, 0)
		for k, v := range detail {
			rt, err := jsonhelper.GenResourceOutput(k, v)
			if err != nil {
				exitOnError(err)
			}
			resources = append(resources, rt)
		}
		sort.Slice(resources, func(i, j int) bool {
			return resources[i].Name < resources[j].Name
		})

		output = map[string]interface{}{
			"resources": resources,
		}

		if *diagnosticsOutput {
			output.(map[string]interface{})["diagnostics"] = jsonhelper.GenPortalDiagnosticOutput(covCnt, scmCnt, ignoreUncoveredResources, coverageMap)
		}
	}

	b, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		exitOnError(err)
	}
	fmt.Println(string(b))

}

func exitOnError(err error) {
	log.Println(err.Error())
	os.Exit(1)
}

func diagOutput(covCnt, scmCnt map[string]int, ignoreUncoveredResources *bool, coverageMap map[string]map[string]interface{}) {
	fmt.Println("----------------------------------------")
	totalScm := 0
	totalCov := 0
	// coverage and schema might have difference

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

	issueRes := make([]string, 0)
	fmt.Println("resource coverage detail:")
	for k, _ := range resultCnt {
		percent := float64(covCnt[k]) / float64(scmCnt[k]) * 100
		if covCnt[k] != len(coverageMap[k]) {
			issueRes = append(issueRes, fmt.Sprintf("%s: statics count: %d, coverage count: %d", k, covCnt[k], len(coverageMap[k])))
		}
		fmt.Println(fmt.Sprintf("resource: %s, schema cnt: %d, coverage cnt: %d, percent: %.2f%%", k, scmCnt[k], covCnt[k], percent))
	}
	fmt.Println("----------------------------------------")

	if len(issueRes) > 0 {
		fmt.Println("coverage issue resources:")
		for _, res := range issueRes {
			fmt.Println(res)
		}
	}
	fmt.Println("----------------------------------------")
	fmt.Println(fmt.Sprintf("total resources: %d", len(scmCnt)))
	fmt.Println(fmt.Sprintf("total count schema: %d, coverage: %d, percent: %.2f%%", totalScm, totalCov, float64(totalCov)/float64(totalScm)*100))
	fmt.Println("----------------------------------------")
}
