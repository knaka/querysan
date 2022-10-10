package qsfts

import (
	"context"
	"database/sql"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"log"
)

var dbConn *sql.DB = nil
var ctx context.Context

func OpenDatabase() {
	dbFilePath, err := DbFilePath()
	if err != nil {
		log.Panicf("panic 600b035 (%v)", err)
	}
	dbConn, _ = sql.Open("sqlite3", dbFilePath)
	boil.SetDB(dbConn)
	ctx = context.Background()
}
