# Generic, a code generation tool to enable generics in go.

`go install github.com/taylorchu/generic/cmd/generic`

This is an experiment to enable generics with code generation in the most elegant way.

## Example

You can create a package like this. Note that `Type` and `TypeQueue` are type placeholders.

```go
package queue

type Type string

// TypeQueue represents a queue of Type types.
type TypeQueue struct {
	items []Type
}

// New makes a new empty Type queue.
func New() *TypeQueue {
	return &TypeQueue{items: make([]Type, 0)}
}

// Enq adds an item to the queue.
func (q *TypeQueue) Enq(obj Type) *TypeQueue {
	q.items = append(q.items, obj)
	return q
}

// Deq removes and returns the next item in the queue.
func (q *TypeQueue) Deq() Type {
	obj := q.items[0]
	q.items = q.items[1:]
	return obj
}

// Len gets the current number of Type items in the queue.
func (q *TypeQueue) Len() int {
	return len(q.items)
}
```

This is what a rewrite rule looks like:

```
result Type->int64 TypeQueue->FIFO
```

It says that the __package target__ is `result`, and what each type replaces to.

Put this line in one of your `.go` file.

```go
//go:generate generic github.com/YourName/queue result Type->int64 TypeQueue->FIFO
```

The output is saved to `$PWD/result/` after you run `go generate`.

```go
package result

type FIFO struct {
	items []int64
}

func New() *FIFO {
	return &FIFO{items: make([]int64, 0)}
}

func (q *FIFO) Enq(obj int64) *FIFO {
	q.items = append(q.items, obj)
	return q
}

func (q *FIFO) Deq() int64 {
	obj := q.items[0]
	q.items = q.items[1:]
	return obj
}

func (q *FIFO) Len() int {
	return len(q.items)
}
```

If the type is not a built-in type, for example, `list.Element`, you write rules like this:

`Type->container/list:list.Element`

If the type is from the current package, add a dot `.` to to the beginning of the package target, for example. `.result`.

You can find more examples in `fixture/` and their outputs in `output/`.

## Best practices

### Decide a good package target name

Package target is essentially go package path, and the base name of this package path is a package name.

> Good package names are short and clear. They are lower case, with no under_scores or mixedCaps.

See [Package names](https://blog.golang.org/package-names).

### Replace TypeXXX with types defined in the _current_ package

The package target should start with `.`. For example, `.queue`. The tool will output in the current package (instead of creating a new sub-package) to prevent circular imports. The package name is set from `go generate`.

### Replace TypeXXX with _built-in_ types, or types defined in _another_ package

The package target should start with `internal/`. For example, `internal/queue`.

> When the go command sees an import of a package with internal in its path, it verifies that the package doing the import is within the tree rooted at the parent of the internal directory. For example, a package .../a/b/c/internal/d/e/f can be imported only by code in the directory tree rooted at .../a/b/c. It cannot be imported by code in .../a/b/g or in any other repository.

See [Internal packages](https://golang.org/doc/go1.4#internalpackages).

## Existing approaches to generics in go

  - code generation
    - output
      - file
      - package (*)
    - rewrite method
      - simple string replacement
      - ast-based replacement (*)
    - type placeholder
      - text/template, i.e. `{{ .Type }}`
      - special types from a package, i.e. `generic.Type`
      - in-package type declaration, i.e. `type A int`
        - above-declaration comment, i.e. `// template type Vector(A, N)`
        - without comment
          - any type name
          - type name with certain pattern (*)
  - language change
  - interface{}
  - reflect
  - marshal everything to bytes/string
  - copy&paste

## What does `generic` do?

![](http://i.imgur.com/X07XInF.png)

`generic` does the followings if you put the following comments in your go code:

```go
// generate from a generic package and save result as a new package,
// with a list of rewrite rules!

//go:generate generic github.com/go/sort int Type->int
```

1. Run `go get -u github.com/go/sort` if the package does not exist locally.
  - If the package exists locally, go-get will not be called.
2. Gather `.go` files (skip `_test.go`) in github.com/go/sort
3. Apply AST rewrite to replace Type in those `.go` files to `int`.
  - Only type that starts with __Type__ and is [non-composite](https://golang.org/ref/spec#Types) can be converted. This enables variable naming like __TypeKey__ or __TypeValue__
  that closely expresses meaning while there is still a namespace for type placeholder.
  - Many rewrite rules are possible: `TypeKey->string TypeValue->int`.
  - We can rewrite non-builtin types with `:`: `Type->github.com/go/types:types.Box`.
  - If a type placeholder has methods defined, the replaced type will need to implement those methods.
4. Type-check results.
5. Save the results as a new package called `int` in `$PWD`.
  - If there is already a dir called `int`, it will first be removed.
  - If the new package starts with `.`, it will save the results in `$PWD`:
      - The package name is set to `$GOPACKAGE` from `go generate`.
      - All top-level identifiers will have prefixes to prevent conflicts, and their uses will also be updated.
      - Filenames will be renamed to prevent conflicts.

## FAQ

### Why are type-checking and ast-based replacement important?

Type-checking and ast-based replacement ensure that the tool doesn't generate invalid code even you or the tool make mistakes, and rewrites identifiers in cases that it shouldn't.

### Why is type placeholder designed this way?

`type TypeXXX int32`

 - It provides a namespace for replaceable types.
 - Knowing that this type might be replaced, package creator can still write go-testable code with a concrete type.
 - It can express meaning. For example, `TypeQueue` shows that it is a queue.

### Why does this tool rewrite at package-level instead of file-level?

 - This tool tries NOT to apply any restriction for package creator except that any TypeXXX might be rewritten. Package creator has full flexibility to write normal go code.
 - It is common to distribute go code at package-level.

## LICENSE

The MIT License (MIT)
Copyright (c) 2016 taylorchu.

Permission is hereby granted, free of charge, to any person obtaining a copy of this software and associated documentation files (the "Software"), to deal in the Software without restriction, including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
