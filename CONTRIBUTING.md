# Contributing

Contributions to VPP-Agent are welcome. We use the standard pull request
model. You can either pick an open issue and assign it to yourself or open
a new issue and discuss your feature.

In any case, before submitting your pull request please check the 
[Coding style](CODINGSTYLE.md) and cover the newly added code with tests 
and documentation.

The dependencies are managed using Go modules. On any change of
dependencies, run `go mod tidy` to update `go.mod`/`go.sum` files.