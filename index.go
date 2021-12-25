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
	stmt, err := db.Prepare("INSERT INTO fileinfo(path, title, words) values(?,?,?)")
	if err != nil {
		return err
	}
	title := ""
	content, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	words := words(string(content))
	res, err := stmt.Exec(path, title, strings.Join(words, " "))
	if err != nil {
		return err
	}
	fmt.Println(res)
	return nil
}

func CloseDb() {
	err := db.Close()
	if err != nil {
		log.Fatalln(err)
	}
}
