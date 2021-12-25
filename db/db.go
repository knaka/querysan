package db

import (
	"database/sql"
	"embed"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/knaka/querysan/tokenizer"
	_ "github.com/mattn/go-sqlite3"
)

//go:embed ddl/*.sql
var fsDdl embed.FS

var dbConn *sql.DB

func SimpleConnect(dbPath string) error {
	var err error
	dbConn, err = sql.Open("sqlite3", dbPath)
	if err != nil {
		return err
	}
	return nil
}

func Initialize(dbPath string) error {
	if _, err := os.Stat(dbPath); err != nil {
		if !os.IsNotExist(err) {
			return err
		}
		file, err := os.Create(dbPath)
		if err != nil {
			return err
		}
		if err := file.Close(); err != nil {
			return err
		}
	}
	sourceDriver, err := iofs.New(fsDdl, "ddl")
	if err != nil {
		return err
	}
	databaseUrl := "sqlite3:" + dbPath
	m, err := migrate.NewWithSourceInstance(
		"iofs", sourceDriver,
		databaseUrl,
	)
	if err != nil {
		return err
	}
	if err := m.Up(); !(err == nil || err == migrate.ErrNoChange) {
		return err
	}
	dbConn, err = sql.Open("sqlite3", dbPath)
	if err != nil {
		return err
	}
	return nil
}

func GetPathListUnderDir(dir string) ([]string, error) {
	var paths []string
	rows, err := dbConn.Query("SELECT path FROM fileinfo WHERE path LIKE ?", fmt.Sprintf("%s%c", dir, filepath.Separator)+"%")
	if err != nil {
		return paths, err
	}
	defer func() {
		_ = rows.Close()
	}()
	for rows.Next() {
		var path string
		if err := rows.Scan(&path); err != nil {
			return paths, err
		}
		paths = append(paths, path)
	}
	return paths, nil
}

func Query(query string) ([]string, error) {
	rows, err := dbConn.Query("SELECT path FROM fileinfo WHERE words MATCH ?", query)
	if err != nil {
		return nil, err
	}
	var paths []string
	for rows.Next() {
		var path string
		if err := rows.Scan(&path); err != nil {
			return paths, err
		}
		paths = append(paths, path)
	}
	return paths, nil
}

func insertEntry(path string) error {
	if !strings.HasSuffix(path, "md") {
		return nil
	}
	stmt, err := dbConn.Prepare("INSERT INTO fileinfo(path, title, words, updated_at) values(?, ?, ?, ?)")
	if err != nil {
		return err
	}
	title := ""
	content, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	words := tokenizer.Words(string(content))
	fileInfo, err := os.Stat(path)
	if err != nil {
		return err
	}
	modTime := fileInfo.ModTime()
	updatedAt := toExternalTime(&modTime)
	_, err = stmt.Exec(path, title, strings.Join(words, " "), updatedAt)
	if err != nil {
		return err
	}
	log.Println("Added file:", path)
	return nil
}

func toExternalTime(timeArg *time.Time) string {
	return timeArg.UTC().Format(time.RFC3339Nano)
}

func fromExternalTime(datetime string) (*time.Time, error) {
	// if len(datetime) != 30 {
	// 	return nil, fmt.Errorf("not in rfc3339Nano format: %s", datetime)
	// }
	utcTime, err := time.Parse(time.RFC3339Nano, datetime)
	if err != nil {
		return nil, err
	}
	timeRet := utcTime.Local()
	return &timeRet, nil
}

func deleteEntry(path string) error {
	stmt, err := dbConn.Prepare("DELETE FROM fileinfo WHERE path = ?")
	if err != nil {
		return err
	}
	_, err = stmt.Exec(path)
	if err != nil {
		return err
	}
	log.Println("Removed file:", path)
	return nil
}

func DeleteEntry(path string) error {
	path, err := regulatePath(path)
	if err != nil {
		return err
	}
	return deleteEntry(path)
}

func updateEntry(path string) error {
	if err := deleteEntry(path); err != nil {
		return err
	}
	if err := insertEntry(path); err != nil {
		return err
	}
	return nil
}

func UpsertEntry(path string) error {
	path, err := regulatePath(path)
	if err != nil {
		return err
	}
	row := dbConn.QueryRow("SELECT updated_at FROM fileinfo WHERE path = ?", path)
	var updatedAtExternal string
	if err = row.Scan(&updatedAtExternal); err != nil {
		if err == sql.ErrNoRows {
			return insertEntry(path)
		}
		return err
	}
	updatedAt, err := fromExternalTime(updatedAtExternal)
	if err != nil {
		return err
	}
	fileStat, err := os.Stat(path)
	if err != nil {
		return err
	}
	modTime := fileStat.ModTime()
	if modTime.After(*updatedAt) {
		return updateEntry(path)
	}
	return nil
}

func regulatePath(path string) (string, error) {
	path, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	return filepath.Clean(path), nil
}

func Finalize() {
	err := dbConn.Close()
	if err != nil {
		log.Fatalln(err)
	}
}
