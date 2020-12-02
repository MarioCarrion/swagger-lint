package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"
)

type (
	Resource string
	Verb     string
	Code     string

	Errors map[Resource][]string

	//-

	Schema struct {
		Type string `json:"type"`
		Ref  string `json:"$ref"`
	}

	Parameter struct {
		In     string `json:"in"`
		Schema Schema `json:"schema"`
		Name   string `json:"name"`
	}

	Response struct {
		Schema Schema `json:"schema"`
	}

	Responses map[Code]Response

	//-

	Path struct {
		Tags        []string    `json:"tags"`
		OperationID string      `json:"operationId"`
		Parameters  []Parameter `json:"parameters"`
		Responses   Responses   `json:"responses"`
	}

	Paths map[Verb]Path

	Swagger struct {
		Paths map[Resource]Paths `json:"paths"`
	}
)

func main() {
	var input string

	flag.StringVar(&input, "input", "", "Swagger 2.0 JSON file.")
	flag.Parse()

	file, err := os.Open(input)
	if err != nil {
		log.Fatalf("Reading input: %s", err)
	}
	defer func() {
		if err := file.Close(); err != nil {
			log.Fatalf("Closing input: %s", err)
		}
	}()

	//-

	var res Swagger

	if err := json.NewDecoder(file).Decode(&res); err != nil {
		log.Fatalf("Decoding %s", err)
	}

	errors := res.Validate()

	var count int
	for path, errs := range errors {
		fmt.Printf("\n%s\n", path)
		for _, err := range errs {
			fmt.Printf("\t%s\n", err)
			count++
		}
	}

	if len(errors) > 0 {
		fmt.Printf("\nTotal violations: %d\n", count)
		os.Exit(1)
	}

	fmt.Printf("File Swagger 2.0 linting rules")
}

func (s Swagger) Validate() Errors {
	var res Errors = make(map[Resource][]string)

	for resource, paths := range s.Paths {
		var errors []string

		for verb, path := range paths {
			if path.OperationID == "" {
				errors = append(errors, "Missing operation id.")
			} else if strings.ToLower(path.OperationID)[0:len(verb)] != string(verb) {
				errors = append(errors, fmt.Sprintf("'%s': Resource must begin with '%s'.", path.OperationID, verb))
			}

			if len(path.Tags) == 0 {
				errors = append(errors, "Resource must define at least one tag.")
			}

			prefix := strings.Title(path.OperationID)

			for _, parameter := range path.Parameters {
				if parameter.In == "body" && parameter.Schema.Ref != "" {
					verbAllRegEx := regexp.MustCompile(fmt.Sprintf(`^(#\/definitions\/%sRequest)`, prefix))
					if !verbAllRegEx.MatchString(parameter.Schema.Ref) {
						errors = append(errors, fmt.Sprintf("'%s': Body request model must be prefixed with verb+Request: '%s'.", path.OperationID, parameter.Schema.Ref))
					}
				}

				if parameter.In == "query" && strings.ToLower(parameter.Name) != parameter.Name {
					errors = append(errors, fmt.Sprintf("'%s': Query arguments must be lowercase: '%s'", path.OperationID, parameter.Name))
				}
			}

			for code, response := range path.Responses {
				if strings.HasPrefix(string(code), "2") {
					if response.Schema.Type == "array" {
						errors = append(errors, fmt.Sprintf("'%s': Instead of using Array as a response, prefer definining a new model.", path.OperationID))
					}

					if response.Schema.Ref != "" {
						verbAllRegEx := regexp.MustCompile(fmt.Sprintf(`^(#\/definitions\/%sResponse)`, prefix))
						if !verbAllRegEx.MatchString(response.Schema.Ref) {
							errors = append(errors, fmt.Sprintf("'%s': Code %s, response model must be prefixed with verb+Response: '%s'.", path.OperationID, code, response.Schema.Ref))
						}
					}
				}
			}

			if len(errors) > 0 {
				res[resource] = errors
			}
		}
	}

	return res
}
