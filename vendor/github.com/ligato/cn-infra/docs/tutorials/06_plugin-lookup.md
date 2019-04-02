# Tutorial: Managing plugin dependencies with plugin lookup

In this tutorial, we learn how to resolve dependencies between multiple plugins in various scenarios. 

Requirements:
* Complete and understand the ['Hello World Agent'](01_hello-world.md) tutorial
* Complete and understand the ['Plugin Dependencies'](02_plugin-deps.md) tutorial

The vpp-agent is based on plugins. A plugin is the go structure satisfying the plugin interface defined by the infrastructure. The plugin often performs only specified task (like a connection to the KVDB, starting HTTP handlers, support for a part of the VPP API, etc.) and often requires other plugins to work with. A good example is KVDB sync - a plugin synchronizing events from various data stores but requires one or more KVDB plugins to connect and provide the actual data. 

The agent can consist of many plugins each with multiple dependencies, which creates a structure (not always a tree) of dependencies. Those dependencies need to be initialized in the correct order. The rule of thumb is that the plugin listed as a dependency has to be started first. The plugin interface provides two methods to work with: the `Init()` and the `AfterInit()`.

The `Init()` is mandatory for every plugin and its purpose is to initialize plugin fields or other tasks (prepare channel objects, initialize maps, start watchers). If some task is not possible to perform during the initialization, the plugin can call the `AfterInit()`. These scenarios will be shown in the tutorial.

The tutorial is based on the [HelloWorld](01_hello-world.md) plugin from the Hello World Agent example. The original example contains a single plugin with `main()` preparing and starting the agent. Since this tutorial requires multiple files for better readability and understanding, we do the following changes at the beginning:

1. Create a new go file called `plugin1.go` 
2. Move hello world plugin to it (leave only `main()` in the `main.go`)

Let's start with creating another plugin in a separate file called `plugin2.go`:
```go
// HelloUniverse represents another plugin.
type HelloUniverse struct{}

// String is used to identifying the plugin by giving its name.
func (p *HelloUniverse) String() string {
	return "HelloUniverse"
}

// Init is executed on agent initialization.
func (p *HelloUniverse) Init() error {
	log.Println("Hello Universe!")
	return nil
}

// Close is executed on agent shutdown.
func (p *HelloUniverse) Close() error {
	log.Println("Goodbye Universe!")
	return nil
}
```

Now we have three files: `main.go` with the `main()` method, `plugin1.go` with the `HelloWorld` plugin and `plugin2.go` with the `HelloUniverse` plugin.

Next step is to start our second plugin together with the `HelloWorld`. We need to initialize new plugin in the `main()` and add it to the `Plugins` option (also rename the first plugin to `p1`):
```go
	p1 := new(HelloWorld)
	p2 := new(HelloUniverse)
	a := agent.NewAgent(agent.Plugins(p1, p2))
```

When you start/stop the program now, notice that the plugins are started in the same order as they were put to the `agent.Plugin()` method, and stopped in reversed order. This is very important because it ensures the dependency plugin is started before the superior plugin, and the superior plugin is stopped before its dependency.

Now we create a relation between plugins. In order to test it, we need some method in the `HelloUniverse` which can be called from outside. We create a simple flag which will be set in `Init()` and can be retrieved via `IsCreated` method. The method returns true if the `HelloUniverse` plugin was initialized, false otherwise:
````go
type HelloUniverse struct{
	created bool    // add the flag
}
````

Set the `created` field in the `Init()`:
```go
func (p *HelloUniverse) Init() error {
	log.Println("Hello Universe!")
	p.created = true
	return nil
}
```

Define method to obtain the value:
```go
// IsCreated returns true if the plugin was initialized
func (p *HelloUniverse) IsCreated() bool {
	return p.created
}
```

Now we will use this code in our `HelloWorld` plugin, so we need to set a dependency. Let's say that the world cannot exist without the universe:
```go
type HelloWorld struct{
	Universe *HelloUniverse
}
```

Our `HelloWorld` plugin is now dependent at the `HelloUniverse` plugin. Next step is to use the `IsCreated` in the `HelloWorld` method `Init()`:
```go
if p.Universe == nil || !p.Universe.IsCreated() {
    log.Panic("Our world cannot exist without the universe!")
}
log.Println("Hello World!")
return nil
``` 

The `HelloWorld` plugin verifies that the `HelloUniverse` plugin already exists, otherwise it panics. And if we try to run the agent now, it is exactly what happens. The reason is that our order of plugin initialization is not as it should be, but reversed (remember, `agent.Plugins(p1, p2)`). So what options do we have? 

**Note:** at this point you can follow one or more of options below, but to continue with the tutorial, the third option will be required. 

### 1: Use the second layer of initialization - AfterInit()

The first option is to use the `AfterInit()` method. In case the superior plugin can "wait" for dependency and make use of it later, we can remove it from the `Init()` and move it to the second initialization round. In our scenario, our `HelloWorld` will be created without the universe (let's put the logic aside for a while) and verifies its existence later - from the agent perspective, the result is the same - all the plugins and their dependencies are initialized.

Move the following code from `HelloWorld` `Init()` to the `AfterInit()`:
```go
if p.Universe == nil || !p.Universe.IsCreated() {
    log.Panic("Our world cannot exist without the universe!")
}
```

Our plugin can be started now without panicking. But remember, **in this case the dependent plugin is started before the superior plugin and stopped in reversed order**. It means the dependency will be stopped first - attempt to use it in superior plugin `Close()` may cause panic again. Despite the fact that this approach is working and can be absolutely correct in suitable scenarios, we do not recommend it, since the `AfterInit()` can be used a lot better as we see later.

### 2: Manually order plugins

* If you followed the first approach, please move `IsCreated()` back to `Init()`.

The simplest option in our scenario is to just to manually switch plugins. In the `main()`, switch this code:
```go
a := agent.NewAgent(agent.Plugins(p1, p2))
```

to:

```go
a := agent.NewAgent(agent.Plugins(p2, p1))
```

This ensures that the `HelloUniverse` will be started before the `HelloWorld`, so the dependency plugin will be initialized first (and close second). While this approach is useful for small agents, the disadvantage is that it becomes difficult to manage if there are several plugins with multi-level dependencies. Especially when some change in dependency was introduced, the plugin order can be very hard to update. Because of this, the CN-Infra provides an automatic process which manages and re-orders dependencies itself, called plugin lookup.

### 3: Order dependencies using plugin lookup

* If you followed the first approach, please move `IsCreated()` back to `Init()`.
* If you followed the second approach, please set plugin order back to `agent.Plugins(p1, p2)`.

The plugin lookup is an automatic process sorting plugins according to their dependencies. More theoretical information about the plugin lookup can be read [here](https://github.com/ligato/cn-infra/wiki/Agent-Plugin-Lookup).

In our plugin, we replace the `agent.Plugins()` method with the `agent.AllPlugins()` in order to use the plugin lookup feature. However, **only one plugin is recommended to be listed in the method**. Since all dependencies are found automatically, the method needs the only top-level plugin to initialize the whole agent (but setting more than one is not forbidden).

The best practice is to specify another helper plugin which defines all other plugins (otherwise listed in `agent.Plugins()`) as dependencies. This top-level plugin (we will call it `Agent`) will not specify any inner fields, only external dependencies and plugin methods `Init()` and `Close` will be empty. Let's create it in the `main.go` (where the `HelloWorld` plugin was before):
```go
type Agent struct {
	Hw *HelloWorld
	Hu *HelloUniverse
}

func (p *Agent) Init() error {
	return nil
}

func (p *Agent) Close() error {
	return nil
}

func (p *Agent) String() string {
	return "AgentPlugin"
}
```

We also define a function `New()` (without receiver) which sets inner plugin dependencies and returns an instance of the `Agent` plugin:
```go
func New() *Agent {
	hw := &HelloWorld{}
	hu := &HelloUniverse{}

	hw.Universe = hu

	return &Agent{
		Hw: hw,
		Hu: hu,
	}
}
```

Now, remove the old plugin dependency management from `main()`:
```go
// Delete following code
p1 := new(HelloWorld)
p2 := new(HelloUniverse)

p1.Universe = p2
```

Last step is to provide the `Agent` plugin to the plugin lookup:
```go
a := agent.NewAgent(agent.AllPlugins(New()))
```

The `main.go` looks like this:
```go
func main() {
	a := agent.NewAgent(agent.AllPlugins(New()))

	if err := a.Run(); err != nil {
		log.Fatalln(err)
	}
}
```

The agent can be now successfully started. This approach may look like a big overhead since we need one new plugin, but it can save a lot of trouble when an agent with a complicated dependency tree is required.

**Cross dependencies**

Note: Continuation of the tutorial requires the third option of dependency solving (above) completed.

Now we have two plugins, one dependent on another, and third top-level starter plugin in `main.go`. Our `HelloWorld` plugin calls a simple method `IsCreated` checking whether the universe exists. However this method is not useful at all, so let's turn it to something more practical. Following changes are done in the `HelloUniverse` plugin.

Remove the `IsCreated()` method together with the `created` flag and its assignment in the `Init()`. This is how the plugin should look like now:
```go
type HelloUniverse struct{}

func (p *HelloUniverse) String() string {
	return "HelloUniverse"
}

func (p *HelloUniverse) Init() error {
	log.Println("Hello Universe!")
	return nil
}

func (p *HelloUniverse) Close() error {
	log.Println("Goodbye Universe!")
	return nil
}
```

Add the registry map. Since every universe contains many worlds, the plugin keeps the list of their names and sizes and provides a method to register a new world (aka. one plugin registers itself or some part to another, a very common scenario). Define map in the plugin:
```go
type HelloUniverse struct{
	worlds map[string]int
}
```

Initialize the map in the `Init()`:
```go
func (p *HelloUniverse) Init() error {
	log.Println("Hello Universe!")
	p.worlds = make(map[string]int)
	return nil
}
```

And add the exported method which `HelloWorld` plugin can use to register:
```go
func (p *HelloUniverse) RegisterWorld(name string, size int) {
	p.worlds[name] = size
	log.Printf("World %s (size %d) was registered", name, size)
}
```

Now move to the `HelloWorld` plugin. Remove following code from the `Init()` since the `IsCreated` method does not exist anymore:
```go
if p.Universe == nil || !p.Universe.IsCreated() {
    log.Panic("Our world cannot exist without the universe!")
}
```

Instead, register the world under some name:
```go
func (p *HelloWorld) Init() error {
	log.Println("Hello World!")
	p.Universe.RegisterWorld("world1")
	return nil
}
```

The dependency situation was not changed here, the `HelloUniverse` plugin still must be initialized first or an error occurs since the registration map would not be initialized. Now the code can be built and started. 

In the next step, the `HelloUniverse` manipulates with `HelloWorld` using its methods. This requires `HelloUniverse` to have `HelloWorld` as a dependency, creating cross (or circular) dependencies. Such a case is a bit more complicated, since the order of plugins defined in the top-level plugin matters as well.

Start with the `HelloWorld`. The world needs to be placed somewhere and the `HelloUniverse` plugin decides where it has some free space.

Add new method to the `HelloWorld`:
```go
func (p *HelloWorld) SetPlace(place string) {
	log.Printf("world1 was placed %s", place)
}
```

Now the question is when this method should be called by the `HelloUniverse`. It cannot be called during the `Init()` since the `HelloWorld` does not exist at that point, so we must use the `AfterInit()`. The calling sequence will be:

1. `HelloUniverse.Init()` - initialized required fields (map)
2. `HelloWorld.Init()` - initialized plugin and registered to the `HelloUniverse`
3. `HelloUniverse.AfterInit()` - can now manipulate the `HelloWorld` since it is initialized and registered

Set dependency for `HelloUniverse`:
```go
type HelloUniverse struct{
	worlds map[string]int
	
	World *HelloWorld
}
```

Add the following code to the `HelloUniverse`:
```go
func (p *HelloUniverse) AfterInit() error {
	for name := range p.worlds {
        p.World.SetPlace(name, "<some place>")
    }
    return nil
}
```

Then go to `main.go` function `New()` and add dependency:
```go
hw := &HelloWorld{}
hu := &HelloUniverse{}

hw.Universe = hu
hu.World = hw       // add cross dependency
```

The code can be started now.

The important thing to know here is that such a cross-dependency cannot be fully resolved by the automatic plugin lookup, since here the given plugin implementation defines the plugin order, not the dependency itself as before. Let's have a look at the top-level plugin:
```go
type Agent struct {
	Hw *HelloWorld
	Hu *HelloUniverse
}
```

The top-level plugin uses our two plugins as dependencies. The plugin lookup takes the provided plugin `Agent` and reads all dependencies in the order they are defined. The first plugin read is `HelloWorld`, but this plugin has also a dependency on another plugin, `HelloUniverse`. The `HelloUniverse` has dependency as well (on `HelloWorld`) but this plugin is already known to the plugin lookup so it is skipped. The plugin places the `HelloUniverse` first, then `HelloWorld` and the `Agent` is last. Plugin methods `Init()` and `AfterInit()` respectively will be called in this order.

Now we see that the plugin resolution for cross dependencies is based on the order of incriminated plugins. Quick test - switch the plugin order in the `Agent`:
```go
type Agent struct {
	Hu *HelloUniverse
	Hw *HelloWorld
}
``` 

The agent will end up with an error since according to the resolution key above, the automatic lookup puts `HelloWorld` to the first place which is not correct. The ultimate rule in the cross dependencies is that the plugin which should be started first is placed below all dependent plugins.