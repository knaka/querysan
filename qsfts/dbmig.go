package qsfts

import (
	"embed"
	"fmt"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/sqlite3"
	mig_drv_iofs "github.com/golang-migrate/migrate/v4/source/iofs"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"os"
	"path"
	"path/filepath"
	"runtime"
)

//go:embed ddl/*.sql
var fsDdl embed.FS

func DbFilePath() (string, error) {
	var userCacheDir string
	var err error
	switch runtime.GOOS {
	case "darwin":
		userCacheDir = os.Getenv("XDG_CACHE_HOME")
		if userCacheDir == "" {
			homeDir, err := os.UserHomeDir()
			if err != nil {
				return "", fmt.Errorf("error f84a4b6 (%w)", err)
			}
			userCacheDir = path.Join(homeDir, ".cache")
		}
	default:
		userCacheDir, err = os.UserCacheDir()
		if err != nil {
			return "", fmt.Errorf("error 53bc03b (%w)", err)
		}
	}
	return filepath.Join(userCacheDir, "querysan.sqlite3"), nil
}

type logger struct{}

func (*logger) Verbose() bool {
	return true
}

func (*logger) Printf(format string, v ...interface{}) {
	log.Printf(format, v...)
}

func MigrateDatabase() {
	dbFilePath, err := DbFilePath()

	// // migrate で create はしない？
	// if err != nil {
	// 	log.Panicf("panic fcb54f4 (%v)", err)
	// }
	// dbConn, err = sql.Open("sqlite3", dbFilePath)
	// if err != nil {
	// 	log.Panicf("panic a997d0b (%v)", err)
	// }
	// err = dbConn.Close()
	// if err != nil {
	// 	log.Panicf("panic 6c98454 (%v)", err)
	// }

	if err != nil {
		log.Panicf("panic 5d9385a (%v)", err)
	}
	sourceDriver, err := mig_drv_iofs.New(fsDdl, "ddl")
	if err != nil {
		log.Panicf("panic 82595e9 (%v)", err)
	}
	databaseUrl := "sqlite3://" + dbFilePath
	mig, err := migrate.NewWithSourceInstance("querysan migration", sourceDriver, databaseUrl)
	if err != nil {
		log.Panicf("panic fe581ef (%v)", err)
	}
	mig.Log = &logger{}
	if err != nil {
		log.Panicf("panic 733bc43 (%v)", err)
	}
	defer func() { _, _ = mig.Close() }()
	if err := mig.Up(); !(err == nil || err == migrate.ErrNoChange) {
		log.Panicf("panic 416608d (%v)", err)
	}
}
