package common

import (
	"database/sql"
	"fmt"
	"os"
)

// DB is the application-wide database object
var DB *sql.DB

// ConnectDatabase opens a connection to the SQLite database file
func ConnectDatabase() {
	var err error
	DB, err = sql.Open("sqlite3", Config.Core.DatabaseFile)
	if err != nil {
		fmt.Println("Error loading database file: ", Config.Core.DatabaseFile)
		os.Exit(1)
	}
}
