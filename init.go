package querysan

import (
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
	_ "github.com/golang-migrate/migrate/v4/database/sqlite3"
	_ "github.com/mattn/go-sqlite3"
)

func startFileWatching() error {
	watcher, err := fsnotify.NewWatcher()
	var dirsWatched []string
	walkFunc := func(path string, info fs.FileInfo, err error) error {
		if !info.IsDir() {
			return nil
		}
		dir := path
		dirsWatched = append(dirsWatched, dir)
		err = watcher.Add(dir)
		if err != nil {
			return err
		}
		log.Println("Added to watching list: ", dir)
		return nil
	}
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
				if event.Name[0] != '/' {
					log.Fatal("?: ", event.Name)
				}
				log.Println("event:", event)
				if event.Op&fsnotify.Create > 0 {
					fileInfo, err := os.Stat(event.Name)
					x := fileInfo.ModTime()
					fmt.Println(x.UTC())
					if err != nil {
						log.Fatal(err)
					}
					if fileInfo.IsDir() {
						log.Println("cp1 Adding dir: ", event.Name)
						if err := filepath.Walk(event.Name, walkFunc); err != nil {
							log.Fatal(err)
						}
					} else {
						// todo
						log.Println("Should be added to index: ", event.Name)
					}
				} else if event.Op&fsnotify.Write > 0 {
					// todo
					log.Println("Should be updated: ", event.Name)
				} else if event.Op&fsnotify.Remove > 0 {
					log.Println("removed file or directory:", event.Name)
					var dirsWatchedNew []string
					for _, dir := range dirsWatched {
						if strings.HasPrefix(dir, event.Name) {
							if err := watcher.Remove(dir); err != nil {
								log.Println(err)
							}
							log.Println("Removed from watching list: ", dir)
						} else {
							dirsWatchedNew = append(dirsWatchedNew, dir)
						}
					}
					dirsWatched = dirsWatchedNew
					// todo
					log.Println("Should remove index under: ", event.Name)
				} else {
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Println("error:", err)
			}
		}
	}()
	log.Println("cp2 Adding dir")
	if err := filepath.Walk("/Users/knaka/tmp", walkFunc); err != nil {
		return err
	}
	<-done
	return nil
}

func Initialize() error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	dataDir := filepath.Join(homeDir, ".local", "share", "querysan")
	dbPath := filepath.Join(dataDir, "main.db")
	if err = os.MkdirAll(dataDir, os.ModePerm); err != nil {
		return err
	}
	err = InitializeDatabase(dbPath)
	if err != nil {
		return err
	}
	err = startFileWatching()
	if err != nil {
		return err
	}
	return nil
}

func timezone() (string, int) {
	zone, offset := time.Now().Zone()
	return zone, offset
}
