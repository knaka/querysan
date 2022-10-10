package main

import (
	"changeme/qsfts"
	"context"
	"fmt"
	"log"
)

// App struct
type App struct {
	ctx context.Context
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts. The context is saved,
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx

	var err error
	err = qsfts.EnsureConfigFile()
	if err != nil {
		log.Panicf("panic 33f057b (%v)", err)
	}
	conf, err := qsfts.ReadConfig()
	if err != nil {
		log.Panicf("panic c442d75 (%v)", err)
	}
	log.Println(conf)
}

// Greet returns a greeting for the given name
func (a *App) Greet(name string) string {
	return fmt.Sprintf("Hello %s, It's show time!", name)
}
