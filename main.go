package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"regexp"
	"strings"
)

type (
	Resource string
	Method   string
	Code     string

	Violations map[Resource][]string

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

	Path struct {
		Tags        []string    `json:"tags"`
		OperationID string      `json:"operationId"`
		Parameters  []Parameter `json:"parameters"`
		Responses   Responses   `json:"responses"`
	}

	Paths map[Method]Path

	Swagger struct {
		Paths map[Resource]Paths `json:"paths"`
	}
)

func main() {
	var input string

	flag.StringVar(&input, "input", "", "Swagger 2.0 JSON file.")
	flag.Parse()

	//-

	b, err := ioutil.ReadFile(input)
	if err != nil {
		log.Fatalf("Reading input: %s", err)
	}

	//-

	var res Swagger

	if err := json.NewDecoder(bytes.NewReader(b)).Decode(&res); err != nil {
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

	fmt.Println("File follows the expected Swagger 2.0 rules")
}

func (s Swagger) Validate() Violations {
	var res Violations = make(map[Resource][]string)

	for resource, paths := range s.Paths {
		var violations []string

		for verb, path := range paths {
			if vs := s.validateOperationID(verb, path); len(vs) > 0 {
				violations = append(violations, vs...)
			}

			if vs := s.validateParameters(path); len(vs) > 0 {
				violations = append(violations, vs...)
			}

			if vs := s.validateResponses(path); len(vs) > 0 {
				violations = append(violations, vs...)
			}

			if len(violations) > 0 {
				res[resource] = violations
			}
		}
	}

	return res
}

func (s Swagger) validateOperationID(method Method, path Path) []string {
	res := make([]string, 0, 3)

	if path.OperationID == "" {
		res = append(res, "Missing operation id.")
	} else if strings.ToLower(path.OperationID)[0:len(method)] != string(method) {
		res = append(res, newViolation(path.OperationID, "Resource must begin with '%s'.", method))
	}

	if len(path.Tags) == 0 {
		res = append(res, "Resource must define at least one tag.")
	}

	return res
}

func (s Swagger) validateParameters(path Path) []string {
	res := make([]string, 0, 2)

	for _, parameter := range path.Parameters {
		if parameter.In == "body" && parameter.Schema.Ref != "" {
			if !matchRegEx(parameter.Schema.Ref, `^(#\/definitions\/%sRequest)`, path.OperationID) {
				res = append(res,
					newViolation(path.OperationID,
						"Body request model must be prefixed with method+Request: '%s'.", parameter.Schema.Ref))
			}
		}

		if parameter.In == "query" && strings.ToLower(parameter.Name) != parameter.Name {
			res = append(res,
				newViolation(path.OperationID, "Query arguments must be lowercase: '%s'", parameter.Name))
		}
	}

	return res
}

func (s Swagger) validateResponses(path Path) []string {
	res := make([]string, 0, 2)

	for code, response := range path.Responses {
		if strings.HasPrefix(string(code), "2") {
			if response.Schema.Type == "array" {
				res = append(res,
					newViolation(path.OperationID, "Instead of using Array as a response, prefer defining a new model."))
			}

			if response.Schema.Ref != "" {
				if !matchRegEx(response.Schema.Ref, `^(#\/definitions\/%sResponse)`, path.OperationID) {
					res = append(res,
						newViolation(path.OperationID,
							"Code %s, response model must be prefixed with method+Response: '%s'.", code, response.Schema.Ref))
				}
			}
		}
	}

	return res
}

func matchRegEx(value, format, operationID string) bool {
	prefix := strings.Title(operationID)

	methodRegEx := regexp.MustCompile(fmt.Sprintf(format, prefix))

	return methodRegEx.MatchString(value)
}

func newViolation(operationID string, format string, a ...interface{}) string {
	var prefix string

	if operationID != "" {
		prefix = fmt.Sprintf("'%s': ", operationID)
	}

	return fmt.Sprintf("%s%s", prefix, fmt.Sprintf(format, a...))
}
