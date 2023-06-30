package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"bridge-coverage/jsonhelper"
	"bridge-coverage/runner"
)

func main() {
	bridgeFile := flag.String("input", "", "the input file of schema")
	schemaFile := flag.String("schema", "", "the schema dump of azurerm provider")
	ignoreSchemas := flag.String("ignore-schema", "", "the schema to ignore of azurerm provider")
	mapIdentity := flag.String("map-identity", "0", "identity(key) for Element in TypeMap")

	flag.Parse()

	bridgeMap, err := jsonhelper.ParseBridgeFile(*bridgeFile)
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

	defaultMapIdentity := "0"
	if mapIdentity == nil || *mapIdentity == "" {
		mapIdentity = &defaultMapIdentity
	}

	r, err := runner.NwRunner(runner.Opts{
		Resources:     schema.ProviderSchema.ResourcesMap,
		BridgeMap:     bridgeMap,
		IgnoreSchemas: ignoreSchemaList,
		MapIdentity:   *mapIdentity,
	})
	if err != nil {
		exitOnError(err)
	}

	detail, scmCnt, covCnt, err := r.Run()
	if err != nil {
		exitOnError(err)
	}

	b, err := json.MarshalIndent(detail, "", "  ")
	if err != nil {
		exitOnError(err)
	}

	fmt.Println(string(b))

	totalScm := 0
	totalCov := 0
	// schema is the superset of coverage
	for k, c := range scmCnt {
		totalScm += c
		totalCov += covCnt[k]
		percent := float64(covCnt[k]) / float64(c) * 100
		fmt.Println(fmt.Sprintf("resource: %s, schema cnt: %d, coverage cnt: %d, percent: %.2f%%", k, scmCnt[k], covCnt[k], percent))
	}

	fmt.Println("total schema count: ", totalScm)
	fmt.Println("total coverage count: ", totalCov)
	fmt.Println(fmt.Sprintf("total coverage percent: %.2f%%", float64(totalCov)/float64(totalScm)*100))
}

func exitOnError(err error) {
	log.Println(err.Error())
	os.Exit(1)
}
