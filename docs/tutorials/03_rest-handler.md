# Tutorial: REST handler

In this tutorial we will learn how to add custom handler to REST plugin.

The REST plugin provides methods for registering custom HTTP handlers. The REST
plugin API is defined as:

```go
type HandlerProvider func(formatter *render.Render) http.HandlerFunc

type HTTPHandlers interface {
	RegisterHTTPHandler(path string, provider HandlerProvider, methods ...string) *mux.Route
	// ...
}
```

To use the REST plugin we simply define field for it in our plugin.

```go
type MyPlugin struct {
	infra.PluginDeps
	REST rest.HTTPHandlers
}
```

And now we can set the instance in our constructor. Most of plugins have 
default plugin instance define as a global variable that can be used.

```go
func NewMyPlugin() *MyPlugin {
	// ...
	p.REST = &rest.DefaultPlugin
	return p
}
```

Now we define our handler.

```go
func (p *MyPlugin) fooHandler(formatter *render.Render) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			p.Log.Errorf("Error reading body: %v", err)
			http.Error(w, "can't read body", http.StatusBadRequest)
			return
		}
		formatter.Text(w, http.StatusOK, fmt.Sprintf("Hello %s", body))
	}
}
```

And register it to the REST plugin in the Init method.

```go
func (p *MyPlugin) Init() error {
	// ...
	p.REST.RegisterHTTPHandler("/greeting", p.fooHandler, "POST")
	return nil
}
```

Now we run the app and try using `curl`: 

```sh
curl -X POST -d 'John' -H "Content-Type: application/json" http://localhost:9191/greeting
// outputs: Hello John
```

Complete working example can be found at [examples/tutorials/03_rest-handler](../../examples/tutorials/03_rest-handler).
