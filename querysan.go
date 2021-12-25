package querysan

import (
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/knaka/querysan/db"

	"github.com/fsnotify/fsnotify"
	_ "github.com/golang-migrate/migrate/v4/database/sqlite3"
	_ "github.com/mattn/go-sqlite3"
)

func startFileWatching() error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	defer func() {
		_ = watcher.Close()
	}()
	var dirsWatched []string
	walkFunc := func(path string, fileInfo fs.FileInfo, err error) error {
		if !fileInfo.IsDir() {
			if err := db.UpsertEntry(path); err != nil {
				return err
			}
			return nil
		}
		dir := path
		err = watcher.Add(dir)
		if err != nil {
			return err
		}
		dirsWatched = append(dirsWatched, dir)
		log.Println("Added dir:", dir)
		return nil
	}
	doneWatcher := make(chan bool)
	doneInitialization := make(chan bool)
	go func() {
		<-doneInitialization
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if event.Name[0] != '/' {
					log.Fatal("?: ", event.Name)
				}
				log.Println("Event:", event)
				if event.Op&fsnotify.Create > 0 {
					fileInfo, err := os.Stat(event.Name)
					if err != nil {
						log.Fatal(err)
					}
					if fileInfo.IsDir() {
						if err := filepath.Walk(event.Name, walkFunc); err != nil {
							log.Fatal(err)
						}
					} else {
						if err := db.UpsertEntry(event.Name); err != nil {
							log.Fatal(err)
						}
					}
				} else if event.Op&fsnotify.Remove > 0 || event.Op&fsnotify.Rename > 0 {
					// 消されたのは directory と想定。処理には改善の余地あり
					var dirsWatchedNew []string
					dirRemoved := false
					for _, dir := range dirsWatched {
						if dir == event.Name {
							if err := DeleteRemovedFileEntriesUnderDir(dir); err != nil {
								log.Fatal(err)
							}
							dirRemoved = true
							// Watcher has removed internally
							log.Println("Removed parent dir:", dir)
						} else if strings.HasPrefix(dir, fmt.Sprintf("%s%c", event.Name, filepath.Separator)) {
							if err := watcher.Remove(dir); err != nil {
								log.Fatal(err)
							}
							log.Println("Removed child dir:", dir)
						} else {
							dirsWatchedNew = append(dirsWatchedNew, dir)
						}
					}
					dirsWatched = dirsWatchedNew
					// 消されたのはファイルと想定
					if !dirRemoved {
						if err := db.DeleteEntry(event.Name); err != nil {
							log.Fatal(event.Name)
						}
					}
				} else if event.Op&fsnotify.Write > 0 {
					if err := db.UpsertEntry(event.Name); err != nil {
						log.Fatal(err)
					}
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					doneWatcher <- true
					return
				}
				log.Println("error:", err)
			}
		}
	}()
	// const pathRoot = "/Users/knaka/tmp/test"
	const pathRoot = "/Users/knaka/Dropbox/doc/2021"
	if err := DeleteRemovedFileEntriesUnderDir(pathRoot); err != nil {
		return err
	}
	// Add all files which exists in the directory to the index
	if err := filepath.Walk(pathRoot, walkFunc); err != nil {
		return err
	}
	doneInitialization <- true
	<-doneWatcher
	return nil
}

func DeleteRemovedFileEntriesUnderDir(dir string) error {
	filePaths, err := db.GetPathListUnderDir(dir)
	if err != nil {
		return err
	}
	for _, filePath := range filePaths {
		_, err := os.Stat(filePath)
		if os.IsNotExist(err) {
			if err := db.DeleteEntry(filePath); err != nil {
				return err
			}
			continue
		}
		if err != nil {
			return err
		}
	}
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
	err = db.Initialize(dbPath)
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
