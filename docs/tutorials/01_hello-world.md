# Tutorial: Hello World

In this tutorial we will learn how to create simple plugin that prints "Hello World".

The plugin interface is defined as:

```go
type Plugin interface {
	// Init is called in the agent`s startup phase.
	Init() error
	// Close is called in the agent`s cleanup phase.
	Close() error
	// String returns unique name of the plugin.
	String() string
}
```

So first we create define our plugin and the methods required to implement 
the `Plugin` interface from infra package:

```go
type HelloWorld struct{}

func (p *HelloWorld) String() string {
	return "HelloWorld"
}

func (p *HelloWorld) Init() error {
	log.Println("Hello World!")
	return nil
}

func (p *HelloWorld) Close() error {
	log.Println("Goodbye World!")
	return nil
}
```

Optionally, we could define `AfterInit` method which is executed after 
initialization of all other plugins.

The `AfterInit` method comes from `PostInit` interface defined in infra package as:

```go
type PostInit interface {
	// AfterInit is called once Init() of all plugins have returned without error.
	AfterInit() error
}
```

Now, in our main function we can initialize an instance of our plugin 
and create new agent that adds the instance to the list of agent plugins.

```go
func main() {
    	p := new(HelloWorld)    
    	a := agent.NewAgent(agent.Plugins(p))
    	// ...
}
```

The plugins are added to the agent using agent options. The option `agent.Plugins` 
adds single plugin instance to the list of plugins and the option `agent.AllPlugins` 
adds plugin instance along with all of its dependencies.

Now we can start the agent using `Run()` method which will initialize all of its
plugins by calling their `Init` and `AfterInit` methods, then wait until interrupt
and then stops all of this plugins by calling their `Close` methods.

```go
if err := a.Run(); err != nil {
	log.Fatalln(err)
}
```

Complete working example can be found at [examples/tutorials/01_hello-world](../../examples/tutorials/01_hello-world).
