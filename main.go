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

	r := runner.NwRunner(
		schema.ProviderSchema.ResourcesMap,
		bridgeMap,
		ignoreSchemaList,
	)

	detail, scmCnt, covCnt, err := r.Run()
	if err != nil {
		exitOnError(err)
	}

	b, err := json.MarshalIndent(detail, "", "  ")
	if err != nil {
		exitOnError(err)
	}

	totalScm := 0
	for k, c := range scmCnt {
		totalScm += c
		fmt.Println(fmt.Sprintf("resource: %s, schema cnt: %d, coverage cnt: %d", k, scmCnt[k], covCnt[k]))
	}

	totalCov := 0
	for _, c := range covCnt {
		totalCov += c
	}

	fmt.Println(string(b))
	fmt.Println("total schema count: ", totalScm)
	fmt.Println("total coverage count: ", totalCov)
}

func exitOnError(err error) {
	log.Println(err.Error())
	os.Exit(1)
}
