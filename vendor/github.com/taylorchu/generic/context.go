package generic

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
)

// Context specifies rewrite inputs.
type Context struct {
	FromPkgPath string
	PkgPath     string
	PkgName     string
	Local       bool
	TypeMap     map[string]Target
}

// NewContext creates a new rewrite context.
func NewContext(pkgPath, newPkgPath string, rules ...string) (*Context, error) {
	ctx := &Context{
		FromPkgPath: pkgPath,
	}
	if strings.HasPrefix(newPkgPath, ".") {
		ctx.Local = true
		ctx.PkgPath = strings.TrimPrefix(newPkgPath, ".")
		ctx.PkgName = os.Getenv("GOPACKAGE")
		if ctx.PkgName == "" {
			return nil, errors.New("GOPACKAGE cannot be empty")
		}
	} else {
		ctx.PkgPath = newPkgPath
		ctx.PkgName = filepath.Base(newPkgPath)
	}

	typeMap, err := ParseTypeMap(rules)
	if err != nil {
		return nil, err
	}
	ctx.TypeMap = typeMap

	return ctx, nil
}
