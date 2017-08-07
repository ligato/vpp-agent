package importer

import (
	"go/ast"
	"go/build"
	"go/parser"
	"go/token"
	"go/types"
	"path/filepath"
)

type importer struct {
	cache map[string]*types.Package
}

// New creates a new types.Importer.
//
// See https://github.com/golang/go/issues/11415.
// Many applications use the gcimporter package to read type information from compiled object files.
// There's no guarantee that those files are even remotely recent.
func New() types.ImporterFrom {
	return &importer{
		cache: make(map[string]*types.Package),
	}
}

func (i *importer) Import(pkgPath string) (*types.Package, error) {
	return i.ImportFrom(pkgPath, "", 0)
}

func (i *importer) ImportFrom(pkgPath string, srcDir string, _ types.ImportMode) (*types.Package, error) {
	if pkgPath == "unsafe" {
		return types.Unsafe, nil
	}

	if pkg, ok := i.cache[pkgPath]; ok {
		return pkg, nil
	}

	ctx := build.Default
	ctx.CgoEnabled = false
	buildP, err := ctx.Import(pkgPath, srcDir, 0)
	if err != nil {
		return nil, err
	}

	fset := token.NewFileSet()
	var files []*ast.File
	for _, file := range buildP.GoFiles {
		f, err := parser.ParseFile(fset, filepath.Join(buildP.Dir, file), nil, 0)
		if err != nil {
			return nil, err
		}
		files = append(files, f)
	}

	conf := types.Config{
		Importer: i,
	}

	pkg, err := conf.Check(pkgPath, fset, files, nil)
	if err != nil {
		return nil, err
	}

	i.cache[pkgPath] = pkg
	return pkg, nil
}
