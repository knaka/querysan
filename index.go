package querysan

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
)

//go:embed db/ddl/*.sql
var fsDdl embed.FS

var db *sql.DB

func InitializeDatabase(dbPath string) error {
	if _, err := os.Stat(dbPath); err != nil {
		file, err := os.Create(dbPath)
		if err != nil {
			return err
		}
		if err := file.Close(); err != nil {
			return err
		}
	}
	sourceDriver, err := iofs.New(fsDdl, "db/ddl")
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
	db, err = sql.Open("sqlite3", dbPath)
	if err != nil {
		return err
	}
	return nil
}

func GetIndexPathList() ([]string, error) {
	var paths []string
	rows, err := db.Query("SELECT path FROM fileinfo")
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
	var paths []string
	rows, err := db.Query("SELECT path FROM fileinfo WHERE words MATCH ?", query)
	if err != nil {
		return paths, err
	}
	for rows.Next() {
		var path string
		if err := rows.Scan(&path); err != nil {
			return paths, err
		}
		paths = append(paths, path)
	}
	return paths, nil
}

func AddIndex(path string) error {
	path, err := filepath.Abs(path)
	if err != nil {
		return err
	}
	stmt, err := db.Prepare("INSERT INTO fileinfo(path, title, words, updated_at) values(?, ?, ?, ?)")
	if err != nil {
		return err
	}
	title := ""
	content, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	words := words(string(content))
	fileInfo, err := os.Stat(path)
	if err != nil {
		return err
	}
	modTime := fileInfo.ModTime()
	updatedAt := toExternalDateTime(&modTime)
	_, err = stmt.Exec(path, title, strings.Join(words, " "), updatedAt)
	if err != nil {
		return err
	}
	log.Println("Added:", path)
	return nil
}

const (
	RFC3339Milli = "2006-01-02T15:04:05.999Z07:00"
)

func toExternalDateTime(time1 *time.Time) string {
	return time1.UTC().Format(time.RFC3339Nano)
}

func fromExternalDateTime(datetime string) (*time.Time, error) {
	if len(datetime) != 30 {
		return nil, fmt.Errorf("not in rfc3339Nano format: %s", datetime)
	}
	utcTime, err := time.Parse(time.RFC3339Nano, datetime)
	if err != nil {
		return nil, err
	}
	time1 := utcTime.Local()
	return &time1, nil
}

func RemoveIndex(path string) error {
	path, err := filepath.Abs(path)
	if err != nil {
		return err
	}
	stmt, err := db.Prepare("DELETE FROM fileinfo WHERE path = ?")
	if err != nil {
		return err
	}
	_, err = stmt.Exec(path)
	if err != nil {
		return err
	}
	log.Println("Removed:", path)
	return nil
}

func UpdateIndex(path string) error {
	if err := RemoveIndex(path); err != nil {
		return err
	}
	if err := AddIndex(path); err != nil {
		return err
	}
	return nil
}

func AddOrUpdateIndex(path string) error {
	path, err := filepath.Abs(path)
	if err != nil {
		return err
	}
	row := db.QueryRow("SELECT updated_at FROM fileinfo WHERE path = ?", path)
	var updatedAtExternal string
	if err != nil {
		return err
	}
	if err = row.Scan(&updatedAtExternal); err != nil {
		if err == sql.ErrNoRows {
			return AddIndex(path)
		}
		return err
	}
	updatedAt, err := fromExternalDateTime(updatedAtExternal)
	if err != nil {
		return err
	}
	fileStat, err := os.Stat(path)
	if err != nil {
		return err
	}
	modTime := fileStat.ModTime()
	if modTime != *updatedAt && modTime.Before(*updatedAt) {
		return UpdateIndex(path)
	}
	return nil
}

func CloseDb() {
	err := db.Close()
	if err != nil {
		log.Fatalln(err)
	}
}
