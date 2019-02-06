# Agent [![GoDoc](https://godoc.org/github.com/ligato/cn-infra/agent?status.svg)](https://godoc.org/github.com/ligato/cn-infra/agent)

The **agent** package provides the life-cycle management agent for plugins.
It is intended to be used as a base point of an application used in main package.

```go
func main() {
	plugin := myplugin.NewPlugin()
	
	a := agent.NewAgent(
		agent.Plugins(plugin),
	)
	if err := a.Run(); err != nil {
		log.Fatal(err)
	}
}
```

## Agent options

There are various options available to customize agent:

- `Version(ver, date, id)` sets version of the program
- `QuitOnClose(chan)` sets signal used to quit the running agent when closed
- `QuitSignals(signals)` sets signals used to quit the running agent (default: SIGINT, SIGTERM)
- `StartTimeout(dur)/StopTimeout(dur)` sets start/stop timeout (defaults: 15s/5s)

Add plugins to list of plugins managed by agent using:
- `Plugins(...)` adds just single plugins
- `AllPlugins(...)` adds plugin along with all of its plugin deps
  
See all options [here](https://godoc.org/github.com/ligato/cn-infra/agent#Option).
