package main

import "log"

// HelloWorld represents our plugin.
type HelloWorld struct{
	Universe *HelloUniverse
}

// String is used to identifying the plugin by giving its name.
func (p *HelloWorld) String() string {
	return "HelloWorld"
}

// Init is executed on agent initialization.
func (p *HelloWorld) Init() error {
	log.Println("Hello World!")
	p.Universe.RegisterWorld("Arrakis", 10)
	return nil
}

// AfterInit is executed after initialization of all plugins. It's optional
// and used for executing operations that require plugins to be initialized.
func (p *HelloWorld) AfterInit() error {
	log.Println("All systems go!")
	return nil
}

// Close is executed on agent shutdown.
func (p *HelloWorld) Close() error {
	log.Println("Goodbye World!")
	return nil
}

// SetPlace is an exported method that allows other plugins to set some internal parameters
func (p *HelloWorld) SetPlace(world, place string) {
	log.Printf("%s was placed %s", world, place)
}
