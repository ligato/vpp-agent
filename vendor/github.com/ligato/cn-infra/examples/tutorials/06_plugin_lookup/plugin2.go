package main

import "log"

// HelloUniverse represents another plugin.
type HelloUniverse struct{
	worlds map[string]int

	World *HelloWorld
}

// String is used to identify the plugin by giving it name.
func (p *HelloUniverse) String() string {
	return "HelloUniverse"
}

// Init is executed on agent initialization.
func (p *HelloUniverse) Init() error {
	log.Println("Hello Universe!")
	p.worlds = make(map[string]int)
	return nil
}

// AfterInit is executed after alll plugin's Init()
func (p *HelloUniverse) AfterInit() error {
	for name := range p.worlds {
		p.World.SetPlace(name, "near Canopus")
	}
	return nil
}

// Close is executed on agent shutdown.
func (p *HelloUniverse) Close() error {
	log.Println("Goodbye Universe!")
	return nil
}

// RegisterWorld is exported for other plugins to use
func (p *HelloUniverse) RegisterWorld(name string, size int) {
	p.worlds[name] = size
	log.Printf("World %s (size %d) was registered", name, size)
}
