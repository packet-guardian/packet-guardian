// This source file is part of the Packet Guardian project.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//
// +build dbmysql dball

package common

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/go-sql-driver/mysql" // MySQL driver
)

func init() {
	dbInits["mysql"] = func(d *DatabaseAccessor, c *Config) error {
		var err error
		if c.Database.Port == 0 {
			c.Database.Port = 3306
		}
		mc := &mysql.Config{
			User:              c.Database.Username,
			Passwd:            c.Database.Password,
			Addr:              fmt.Sprintf("%s:%d", c.Database.Address, c.Database.Port),
			DBName:            c.Database.Name,
			Strict:            true,
			InterpolateParams: true,
		}
		d.DB, err = sql.Open("mysql", mc.FormatDSN())
		if err != nil {
			return err
		}

		err = d.DB.Ping()
		if err != nil {
			return err
		}

		d.Driver = "mysql"

		// Check the SQL mode, the user is responsible for setting it
		row := d.DB.QueryRow(`SELECT @@GLOBAL.sql_mode`)

		mode := ""
		if err := row.Scan(&mode); err != nil {
			return err
		}

		ansiOK := strings.Contains(mode, "ANSI")
		tradOK := strings.Contains(mode, "TRADITIONAL")

		if !ansiOK || !tradOK {
			return errors.New("MySQL must be in ANSI,TRADITIONAL mode. Please set the global mode or edit the my.cnf file to enable ANSI,TRADITIONAL sql_mode.")
		}
		return err
	}
}
