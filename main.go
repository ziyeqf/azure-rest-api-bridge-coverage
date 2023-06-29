package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"

	"bridge-coverage/jsonhelper"
	"bridge-coverage/runner"
)

func main() {
	bridgeFile := flag.String("bridge", "", "the output file of output")
	schemaFile := flag.String("schema", "", "the schema file of provider")
	flag.Parse()

	bridgeMap, err := jsonhelper.ParseBridgeFile(*bridgeFile)
	if err != nil {
		log.Println(err.Error())
		os.Exit(1)
	}

	schema, err := jsonhelper.ParseSchema(*schemaFile)
	if err != nil {
		log.Println(err.Error())
		os.Exit(1)
	}

	_ = bridgeMap
	fmt.Println(schema.ProviderSchema.ResourcesMap["azurerm_resource_group"])

	result, err := runner.NwRunner(schema.ProviderSchema.ResourcesMap, bridgeMap).Run()
	if err != nil {
		log.Println(err.Error())
		os.Exit(1)
	}

	b, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		log.Println(err.Error())
		os.Exit(1)
	}

	fmt.Println(string(b))
}
