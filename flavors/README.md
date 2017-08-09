# Flavors

A flavor in this context is a collection of plugins that allows inspecting capabilities available 
to an agent. By convention, flavor provides `Inject()` method that interconnects plugins - injects dependencies and
`Plugins()` method returning all plugins contained in the flavor.

Example:
```go


type Flavor struct {
	injected     bool
	Aplugin      a.Plugin
	Bplugin      b.Plugin
	Cplugin      c.Plugin
}

func (f *Flavor) Inject() error {
	if f.injected {
		return nil
	}
	f.injected = true
	f.Bplugin.A = &f.Aplugin
	f.Cplugin.B = &f.Bplugin
	return nil
}

func (f *Flavor) Plugins() []*core.NamedPlugin {
	f.Inject()
	return core.ListPluginsInFlavor(f)
}

```

This package contains flavors that can be used out-of-the-box.
