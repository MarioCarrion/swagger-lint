# Opinionated Swagger 2.0 Linter

## Installing

`swagger-lint` requires Go 1.15 or greater, install it using:

```
go install github.com/MarioCarrion/swagger-lint
```

For projects depending on `swagger-lint` you could use the [`tools.go` paradigm](https://github.com/go-modules-by-example/index/blob/master/010_tools/README.md):

```go
// +build tools

package tools

import (
	_ "github.com/MarioCarrion/swagger-lint"
)
```

## Using

After installing you can use:

```
swagger-lint -input <full path to swagger.json>
```

## Rules

1. Paths
    * Operation IDs are required.
    * Operation IDs must start with the verb used.
    * At least one extra tag is required.
1. Parameters
    * Body requests models must be named `<operationID>Request`.
    * Query values must be lowercase.
1. Responses (for 2XX status codes only)
    * Response models must be named `<operationID>Response`.
    * Arrays are not encouraged, instead using a new model that follows the rules defined above is preferred.
