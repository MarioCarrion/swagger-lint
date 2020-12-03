package main

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestSwagger_Validate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    Swagger
		expected Violations
	}{
		{
			"Missing Operation ID",
			Swagger{
				Paths: map[Resource]Paths{
					"/items": map[Method]Path{
						"get": {
							Tags: []string{"tag"},
						},
					},
				},
			},
			map[Resource][]string{
				"/items": {"Missing operation id."},
			},
		},
		{
			"Resource must begin with method",
			Swagger{
				Paths: map[Resource]Paths{
					"/items": map[Method]Path{
						"get": {
							OperationID: "itemsOperationID",
							Tags:        []string{"tag"},
						},
					},
				},
			},
			map[Resource][]string{
				"/items": {"'itemsOperationID': Resource must begin with 'get'."},
			},
		},
		{
			"Resource must define at least one tag.",
			Swagger{
				Paths: map[Resource]Paths{
					"/items": map[Method]Path{
						"get": {
							OperationID: "getItems",
						},
					},
				},
			},
			map[Resource][]string{
				"/items": {"Resource must define at least one tag."},
			},
		},
		{
			"Body request model must be prefixed with method+Request",
			Swagger{
				Paths: map[Resource]Paths{
					"/items": map[Method]Path{
						"post": {
							OperationID: "postItems",
							Tags:        []string{"tag"},
							Parameters: []Parameter{
								{
									In: "body",
									Schema: Schema{
										Ref: `#/definitions/postItems`,
									},
								},
							},
						},
					},
				},
			},
			map[Resource][]string{
				"/items": {"'postItems': Body request model must be prefixed with method+Request: '#/definitions/postItems'."},
			},
		},
		{
			"Query arguments must be lowercase",
			Swagger{
				Paths: map[Resource]Paths{
					"/items": map[Method]Path{
						"get": {
							OperationID: "getItems",
							Tags:        []string{"tag"},
							Parameters: []Parameter{
								{
									In:   "query",
									Name: "deletedAt",
								},
							},
						},
					},
				},
			},
			map[Resource][]string{
				"/items": {"'getItems': Query arguments must be lowercase: 'deletedAt'"},
			},
		},
		{
			"Instead of using Array prefer defining a new model.",
			Swagger{
				Paths: map[Resource]Paths{
					"/items": map[Method]Path{
						"get": {
							OperationID: "getItems",
							Tags:        []string{"tag"},
							Responses: map[Code]Response{
								"200": {
									Schema: Schema{
										Type: "array",
									},
								},
							},
						},
					},
				},
			},
			map[Resource][]string{
				"/items": {"'getItems': Instead of using Array as a response, prefer defining a new model."},
			},
		},
		{
			"Code 200, response model must be prefixed with method+Response.",
			Swagger{
				Paths: map[Resource]Paths{
					"/items": map[Method]Path{
						"get": {
							OperationID: "getItems",
							Tags:        []string{"tag"},
							Responses: map[Code]Response{
								"200": {
									Schema: Schema{
										Ref: `#/definitions/getItems`,
									},
								},
							},
						},
					},
				},
			},
			map[Resource][]string{
				"/items": {"'getItems': Code 200, response model must be prefixed with method+Response: '#/definitions/getItems'."},
			},
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			actual := test.input.Validate()

			if !cmp.Equal(test.expected, actual) {
				t.Errorf("expected values do not match\n%s", cmp.Diff(test.expected, actual))
			}
		})
	}
}
