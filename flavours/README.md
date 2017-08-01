# Flavours

A flavour in this context is a collection of plugins that allows inspecting capabilities available 
to an agent. By convention, flavour provides `Inject()` method that interconnects plugins - injects dependencies and
`Plugins()` method returning all plugins contained in the flavour.

Example:
```go


type Flavour struct {
	injected     bool
	Aplugin      a.Plugin
	Bplugin      b.Plugin
	Cplugin      c.Plugin
}

func (f *Flavour) Inject() error {
	if f.injected {
		return nil
	}
	f.injected = true
	f.Bplugin.A = &f.Aplugin
	f.Cplugin.B = &f.Bplugin
	return nil
}

func (f *Flavour) Plugins() []*core.NamedPlugin {
	f.Inject()
	return core.ListPluginsInFlavor(f)
}

```

This package contains flavours that can be used out-of-the-box.
