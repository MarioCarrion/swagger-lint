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

* Paths
    * Operation IDs are required.
    > API users can refer to concrete resources by a using specific IDs, useful when generating code automatically from the final Swagger JSON file.
    * Operation IDs must start with the request method.
    > Allows categorizing operations by the request method being used, similar convention to what [gRPC](https://grpc.io/docs/what-is-grpc/core-concepts/) recommends.
    * Operation IDs should describe the resource. **Not enforced**
    > Allows matching the named resource with the final operation ID. For example: `GET /blogs/{id}/comments` could be named as `getBlogsComment`.
    * At least one tag should be defined.
    > Allows categorizing similar resources.
* Parameters
    * Models for Body requests must be named as `<operationID>Request`.
    > Allows matching the request with the operationID, similar convention to what [gRPC](https://grpc.io/docs/what-is-grpc/core-concepts/) recommends. For example: `PostBlogsRequest`
    * Query values must be lowercase and snake\_case.
    > Self explanatory.
* Responses (for 2XX status codes only)
    * Response models must be named as `<operationID>Response`
    > Allows matching the response with the operationID, similar convention to what [gRPC](https://grpc.io/docs/what-is-grpc/core-concepts/) recommends. For example: `GetBlogsResponse`.
    * Arrays are not encouraged, instead using a new model that follows the rules defined above is preferred.
    > To keep the naming consistent accross the board.
