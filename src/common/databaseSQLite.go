// This source file is part of the Packet Guardian project.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//
// +build dbsqlite dball

package common

import (
	"database/sql"
	"errors"

	_ "github.com/mattn/go-sqlite3" // SQLite driver
)

func init() {
	dbInits["sqlite"] = func(d *DatabaseAccessor, c *Config) error {
		var err error
		if !FileExists(c.Database.Address) {
			return errors.New("SQLite database file doesn't exist")
		}
		d.DB, err = sql.Open("sqlite3", c.Database.Address)
		if err != nil {
			return err
		}

		err = d.DB.Ping()
		if err != nil {
			return err
		}

		d.Driver = "sqlite"
		_, err = d.Exec("PRAGMA foreign_keys = ON")
		return err
	}
}
