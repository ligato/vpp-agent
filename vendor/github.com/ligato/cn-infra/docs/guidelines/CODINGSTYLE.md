# Coding style

# General Rules
- Use `gofmt` to format all source files.
- Address any issues that were discovered by the `golint` & `govet` tool.
- Follow recommendations in [effective go](https://golang.org/doc/effective_go.html) and [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments).
- Please make sure that each dependency in the `glide.yaml` has a specific `version` defined (a specific commit ID or a git tag).

# Go Channels & go routines


# Func signature
## Arguments
Use always name for primitive type if there is more than one return argument (in structure and also in interface).

Correct:
```
func Method() (found bool, data interface{}) {
    // method body ...
}
```
 
Wrong:
```
func Method() (bool, interface{}) {
    // method body ...
}
```
 