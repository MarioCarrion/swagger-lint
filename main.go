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

	//-

	Schema struct {
		Type                 string               `json:"type,omitempty"`
		Items                map[string]string    `json:"items,omitempty"`
		Ref                  string               `json:"$ref,omitempty"`
		AdditionalProperties AdditionalProperties `json:"additionalProperties"`
	}

	AdditionalProperties struct {
		Ref string `json:"$ref,omitempty"`
	}

	Response struct {
		Schema Schema `json:"schema,omitempty"`
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

	for resource, paths := range res.Paths {
		for verb, path := range paths {
			if path.OperationID == "" {
				errors = append(errors, fmt.Sprintf("Path '%s': is missing operation id.", resource))
			} else if strings.ToLower(path.OperationID)[0:len(verb)] != string(verb) {
				errors = append(errors, fmt.Sprintf("Operation ID '%s' Resource '%s': must begin with '%s'.", path.OperationID, resource, verb))
			}

			if len(path.Tags) == 0 {
				errors = append(errors, fmt.Sprintf("Resource '%s': must define at least one tag.", resource))
			}

			for code, response := range path.Responses {
				if strings.HasPrefix(string(code), "2") {
					if response.Schema.Type == "array" {
						errors = validateArray(errors, response.Schema.Items, path.OperationID)
					}
					continue

					verbAllRegEx := regexp.MustCompile(`(?i)^(#\/definitions\/)(post|get|put|delete)`)
					if !verbAllRegEx.MatchString(response.Schema.Ref) {
						errors = append(errors, fmt.Sprintf("Operation ID '%s' Code %s, response model must be prefixed with verb: '%s'", code, path.OperationID, response.Schema.Ref))
					}
				}

				if code == "301" { // redirect
					continue
				}

				resRegEx := regexp.MustCompile(`(?i)response$`)
				if !resRegEx.MatchString(response.Schema.Ref) {
					errors = append(errors, fmt.Sprintf("Operation ID '%s' Code %s, response model must be postfixed with response: '%s'", code, path.OperationID, response.Schema.Ref))
				}
			}
		}
	}

	//-

	PrettyPrint(&res)

	for _, err := range errors {
		fmt.Println(err)
	}

	if len(errors) > 0 {
		os.Exit(1)
	}

}

func validateArray(errors []string, items map[string]string, operationID string) []string {
	v, ok := items["$ref"]
	if !ok {
		errors = append(errors, fmt.Sprintf("Operation ID '%s', Array: is missing the $ref field", operationID))
	}

	resBothRegEx := regexp.MustCompile(`(?i)(response|request)$`)
	if resBothRegEx.MatchString(v) {
		errors = append(errors, fmt.Sprintf("Operation ID '%s', Array: model must be not postfixed with response/request: '%s'", operationID, v))
	}

	verbAllRegEx := regexp.MustCompile(`(?i)^(#\/definitions\/)(post|get|put|delete)`)
	if verbAllRegEx.MatchString(v) {
		errors = append(errors, fmt.Sprintf("Operation ID '%s', Array: model must be not prefixed with verbs: '%s'", operationID, v))
	}

	return errors
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
