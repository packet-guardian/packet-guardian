// This source file is part of the Packet Guardian project.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//
// +build dbmysql dball

package common

import (
	"database/sql"
	"fmt"

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
		_, err = d.DB.Exec(`SET GLOBAL sql_mode="ANSI,TRADITIONAL"`)
		return err
	}
}
