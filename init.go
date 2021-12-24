package querysan

import (
	"embed"
	"log"
	"os"
	"path/filepath"

	"github.com/fsnotify/fsnotify"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/sqlite3"
	miofs "github.com/golang-migrate/migrate/v4/source/iofs"
	_ "github.com/mattn/go-sqlite3"
)

var dataDir string
var dbFile string

//go:embed db/ddl/*.sql
var fs embed.FS

func Initialize() error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	dataDir = filepath.Join(homeDir, ".local", "share", "querysan")
	dbFile = filepath.Join(dataDir, "main.db")
	if _, err = os.Stat(dbFile); err != nil {
		_, err = os.Create(dbFile)
		if err != nil {
			return err
		}
	}
	if err = os.MkdirAll(dataDir, os.ModePerm); err != nil {
		return err
	}
	// err = watch()
	// if err != nil {
	// 	return err
	// }
	err = initDb()
	if err != nil {
		return err
	}
	return nil
}

func watch() error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		_ = watcher.Close()
	}()
	done := make(chan bool)
	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				log.Println("event:", event)
				if filepath.Base(event.Name) == "done" {
					done <- true
				}
				if event.Op&fsnotify.Write == fsnotify.Write {
					log.Println("modified file:", event.Name)
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Println("error:", err)
			}
		}
	}()
	err = watcher.Add("/Users/knaka/tmp")
	if err != nil {
		log.Fatal(err)
	}
	<-done
	return nil
}

func initDb() error {
	sourceDriver, err := miofs.New(fs, "db/ddl")
	databaseUrl := "sqlite3:" + dbFile
	m, err := migrate.NewWithSourceInstance(
		"iofs", sourceDriver,
		databaseUrl,
	)
	if err != nil {
		return err
	}
	err = m.Up()
	if err != nil {
		return err
	}
	return nil
}
