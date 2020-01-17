package models

import (
	"fmt"
	"regexp"
	"strings"

	api "go.ligato.io/vpp-agent/v3/proto/ligato/generic"
)

var (
	validModule  = regexp.MustCompile(`^[-a-z0-9_]+(?:\.[-a-z0-9_]+)?$`)
	validVersion = regexp.MustCompile(`^v[0-9]+(?:[-a-z0-9]+)?$`)
	validType    = regexp.MustCompile(`^[-a-z0-9_]+(?:\.[-a-z0-9_]+)?$`)
	validClass   = regexp.MustCompile(`^[-a-z0-9_]+$`)
)

// Spec defines model specification used for registering model.
type Spec struct {
	Module  string
	Version string
	Type    string
	Class   string
}

func ToSpec(s *api.ModelSpec) Spec {
	return Spec{
		Module:  s.GetModule(),
		Version: s.GetVersion(),
		Type:    s.GetType(),
		Class:   s.GetClass(),
	}
}

func (spec Spec) Proto() *api.ModelSpec {
	return &api.ModelSpec{
		Module:  spec.Module,
		Version: spec.Version,
		Type:    spec.Type,
		Class:   spec.Class,
	}
}

func (spec Spec) KeyPrefix() string {
	modulePath := strings.Replace(spec.Module, ".", "/", -1)
	typePath := strings.Replace(spec.Type, ".", "/", -1)
	return fmt.Sprintf("%s/%s/%s/%s/", spec.Class, modulePath, spec.Version, typePath)
}

func (spec Spec) ModelName() string {
	return fmt.Sprintf("%s.%s", spec.Module, spec.Type)
}

// Validate validates Spec fields.
func (spec Spec) Validate() error {
	if !validModule.MatchString(spec.Module) {
		return fmt.Errorf("invalid module: %q", spec.Module)
	}
	if !validVersion.MatchString(spec.Version) {
		return fmt.Errorf("invalid version: %q", spec.Version)
	}
	if !validType.MatchString(spec.Type) {
		return fmt.Errorf("invalid type: %q", spec.Type)
	}
	if !validClass.MatchString(spec.Class) {
		return fmt.Errorf("invalid class: %q", spec.Class)
	}
	return nil
}
