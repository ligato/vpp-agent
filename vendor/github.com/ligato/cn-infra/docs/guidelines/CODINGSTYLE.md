# Coding style

# General Rules
- Use `gofmt` to format all source files.
- Address any issues that were discovered by the `golint` & `govet` tools.
- Follow recommendations in [effective go][1] and [Go Code Review Comments][2].
- Make sure that each dependency in `glide.yaml` has a specific `version` 
  defined (a specific commit ID or a git tag).

# Go Channels & go routines
See [Plugin Lifecycle](PLUGIN_LIFECYCLE.md)

# Func signature
## Arguments
If there is more than one return argument (in a structure or in an interface),
always use a name for each primitive type.

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
 
[1]: https://golang.org/doc/effective_go.html
[2]: https://github.com/golang/go/wiki/CodeReviewComments
