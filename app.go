package main

import (
	"context"
	"fmt"
	"github.com/knaka/querysan/qsfts"
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
	qsfts.MigrateDatabase()
	err = qsfts.EnsureConfigFile()
	if err != nil {
		log.Panicf("panic 33f057b (%v)", err)
	}
	conf, err := qsfts.ReadConfig()
	if err != nil {
		log.Panicf("panic c442d75 (%v)", err)
	}
	qsfts.OpenDatabase()
	go func() {
		for _, documentDirectory := range conf.DocumentDirectories {
			err := qsfts.ScanFiles(documentDirectory.Path)
			if err != nil {
				log.Panicf("panic 91ff196 (%v)", err)
			}
		}
	}()
	go func() {
		err := qsfts.WatchFiles()
		if err != nil {
			log.Panicf("panic 5225c46 (%v)", err)
		}
	}()
}

// Greet returns a greeting for the given name
func (a *App) Greet(name string) string {
	return fmt.Sprintf("Hello %s, It's show time!", name)
}

func (a *App) Query(query string) []*qsfts.QueryResult {
	return qsfts.Query(query)
}
