# Coding style

- Use `gofmt` to format all source files.
- Address any issues that were discovered by the `golint` & `govet` tool.
- Follow recommendations in [effective go][1] and [Go Code Review Comments][2].
- Please make sure that each dependency in the `glide.yaml` has a specific 
 `version` defined (a specific commit ID or a git tag).

[1]: https://golang.org/doc/effective_go.html
[2]: https://github.com/golang/go/wiki/CodeReviewComments