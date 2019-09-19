package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
)

type (
	Resource string
	Verb     string
	Code     string

	//-

	Schema struct {
		Type  string            `json:"type"`
		Items map[string]string `json:"items"`
		Ref   string            `json:"$ref"`
	}

	Response struct {
		Schema Schema `json:"schema"`
	}

	Responses map[Code]Response

	//-

	Path struct {
		Tags        []string  `json:"tags"`
		OperationID string    `json:"operationId"`
		Responses   Responses `json:"responses"`
	}

	Paths map[Verb]Path

	Swagger struct {
		Paths map[Resource]Paths `json:"paths"`
	}
)

func main() {
	var input string

	flag.StringVar(&input, "input", "", "Swagger JSON file.")
	flag.Parse()

	file, err := os.Open(input)
	if err != nil {
		log.Fatalf("reading file: %s", err)
	}
	defer func() {
		if err := file.Close(); err != nil {
			log.Fatalf("Closing input: %s", err)
		}
	}()

	//-

	var res Swagger
	dec := json.NewDecoder(file)
	if err := dec.Decode(&res); err != nil {
		log.Fatalf("Decoding %s", err)
	}

	//-

	var errors []string

	//1) OperationIDs must be prefixed with the verb value
	for resource, paths := range res.Paths {
		for verb, path := range paths {
			if path.OperationID == "" {
				errors = append(errors, fmt.Sprintf("'%s' is missing operationId", resource))
			} else if strings.ToLower(path.OperationID)[0:len(verb)] != string(verb) {
				errors = append(errors, fmt.Sprintf("'%s' operationId must begin with %s", resource, verb))
			}

			if len(path.Tags) == 0 {
				errors = append(errors, fmt.Sprintf("'%s' must defined at least one tag", resource))
			}
		}
	}

	//-
	for _, err := range errors {
		fmt.Println(err)
	}

	if len(errors) > 0 {
		os.Exit(1)
	}

	// PrettyPrint(&res)
}

// print the contents of the obj
func PrettyPrint(data interface{}) {
	var p []byte
	//    var err := error
	p, err := json.MarshalIndent(data, "", "\t")
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("%s \n", p)
}
