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
	"github.com/lfkeitel/verbose"
)

var mysqlMigrations = []func(*DatabaseAccessor) error{
	1: migrate1to2MySQL,
}

func init() {
	dbInits["mysql"] = initMySQL
}

func initMySQL(d *DatabaseAccessor, c *Config) error {
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

	if !ansiOK {
		return errors.New("MySQL must be in ANSI mode. Please set the global mode or edit the my.cnf file to enable ANSI sql_mode.")
	}

	rows, err := d.DB.Query(`SHOW TABLES`)
	if err != nil {
		return err
	}
	defer rows.Close()
	tables := make(map[string]bool)
	for _, table := range DatabaseTableNames {
		tables[table] = false
	}

	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err != nil {
			return err
		}
		tables[tableName] = true
	}

	if !tables["blacklist"] {
		if err := createMySQLBlacklistTable(d); err != nil {
			return err
		}
	}
	if !tables["device"] {
		if err := createMySQLDeviceTable(d); err != nil {
			return err
		}
	}
	if !tables["lease"] {
		if err := createMySQLLeaseTable(d); err != nil {
			return err
		}
	}
	if !tables["settings"] {
		if err := createMySQLSettingTable(d); err != nil {
			return err
		}
	}
	if !tables["user"] {
		if err := createMySQLUserTable(d); err != nil {
			return err
		}
	}
	if !tables["lease_history"] {
		if err := createMySQLLeaseHistoryTable(d); err != nil {
			return err
		}
	}

	var currDBVer int
	verRow := d.DB.QueryRow(`SELECT "value" FROM "settings" WHERE "id" = 'db_version'`)
	if verRow == nil {
		return errors.New("Failed to get database version")
	}
	verRow.Scan(&currDBVer)

	SystemLogger.WithFields(verbose.Fields{
		"current-version": currDBVer,
		"active-version":  dbVersion,
	}).Debug("Database Versions")

	// No migration needed
	if currDBVer == dbVersion {
		return nil
	}

	neededMigrations := mysqlMigrations[currDBVer:dbVersion]
	for _, migrate := range neededMigrations {
		if migrate == nil {
			continue
		}
		if err := migrate(d); err != nil {
			return err
		}
	}

	_, err = d.DB.Exec(`UPDATE "settings" SET "value" = ? WHERE "id" = 'db_version'`, dbVersion)
	return err
}

func createMySQLBlacklistTable(d *DatabaseAccessor) error {
	sql := `CREATE TABLE "blacklist" (
	    "id" INTEGER PRIMARY KEY AUTO_INCREMENT NOT NULL,
	    "value" VARCHAR(255) NOT NULL UNIQUE KEY,
	    "comment" TEXT
	) ENGINE=InnoDB DEFAULT CHARSET=utf8 AUTO_INCREMENT=1`

	_, err := d.DB.Exec(sql)
	return err
}

func createMySQLDeviceTable(d *DatabaseAccessor) error {
	sql := `CREATE TABLE "device" (
	    "id" INTEGER PRIMARY KEY AUTO_INCREMENT,
	    "mac" VARCHAR(17) NOT NULL UNIQUE KEY,
	    "username" VARCHAR(255) NOT NULL,
	    "registered_from" VARCHAR(15),
	    "platform" TEXT,
	    "expires" INTEGER DEFAULT 0,
	    "date_registered" INTEGER NOT NULL,
	    "user_agent" TEXT,
	    "blacklisted" TINYINT DEFAULT 0,
	    "description" TEXT,
	    "last_seen" INTEGER NOT NULL
	) ENGINE=InnoDB DEFAULT CHARSET=utf8 AUTO_INCREMENT=1`

	_, err := d.DB.Exec(sql)
	return err
}

func createMySQLLeaseTable(d *DatabaseAccessor) error {
	sql := `CREATE TABLE "lease" (
	    "id" INTEGER PRIMARY KEY AUTO_INCREMENT NOT NULL,
	    "ip" VARCHAR(15) NOT NULL UNIQUE KEY,
	    "mac" VARCHAR(17) NOT NULL,
	    "network" TEXT NOT NULL,
	    "start" INTEGER NOT NULL,
	    "end" INTEGER NOT NULL,
	    "hostname" TEXT NOT NULL,
	    "abandoned" TINYINT DEFAULT 0,
	    "registered" TINYINT DEFAULT 0
	) ENGINE=InnoDB DEFAULT CHARSET=utf8 AUTO_INCREMENT=1`

	_, err := d.DB.Exec(sql)
	return err
}

func createMySQLLeaseHistoryTable(d *DatabaseAccessor) error {
	sql := `CREATE TABLE "lease_history" (
	    "id" INTEGER PRIMARY KEY AUTO_INCREMENT NOT NULL,
	    "ip" VARCHAR(15) NOT NULL,
	    "mac" VARCHAR(17) NOT NULL,
	    "network" TEXT NOT NULL,
	    "start" INTEGER NOT NULL,
	    "end" INTEGER NOT NULL
	) ENGINE=InnoDB DEFAULT CHARSET=utf8 AUTO_INCREMENT=1`

	_, err := d.DB.Exec(sql)
	return err
}

func createMySQLSettingTable(d *DatabaseAccessor) error {
	sql := `CREATE TABLE "settings" (
	    "id" VARCHAR(255) PRIMARY KEY NOT NULL,
	    "value" TEXT
	) ENGINE=InnoDB DEFAULT CHARSET=utf8`

	if _, err := d.DB.Exec(sql); err != nil {
		return err
	}

	_, err := d.DB.Exec(`INSERT INTO "settings" ("id", "value") VALUES ('db_version', ?)`, dbVersion)
	return err
}

func createMySQLUserTable(d *DatabaseAccessor) error {
	sql := `CREATE TABLE "user" (
	    "id" INTEGER PRIMARY KEY AUTO_INCREMENT NOT NULL,
	    "username" VARCHAR(255) NOT NULL UNIQUE KEY,
	    "password" TEXT,
	    "device_limit" INTEGER DEFAULT -1,
	    "default_expiration" INTEGER DEFAULT 0,
	    "expiration_type" TINYINT DEFAULT 1,
	    "can_manage" TINYINT DEFAULT 1,
	    "can_autoreg" TINYINT DEFAULT 1,
	    "valid_start" INTEGER DEFAULT 0,
	    "valid_end" INTEGER DEFAULT 0,
	    "valid_forever" TINYINT DEFAULT 1
	) ENGINE=InnoDB DEFAULT CHARSET=utf8 AUTO_INCREMENT=4;`

	if _, err := d.DB.Exec(sql); err != nil {
		return err
	}

	_, err := d.DB.Exec(`INSERT INTO "user"
			("id", "username", "password") VALUES
			(1, 'admin', '$2a$10$rZfN/gdXZdGYyLtUb6LF.eHOraDes3ibBECmWic2I3SocMC0L2Lxa'),
			(2, 'helpdesk', '$2a$10$ICCdq/OyZBBoNPTRmfgntOnujD6INGv7ZAtA/Xq6JIdRMO65xCuNC'),
			(3, 'readonly', '$2a$10$02NG6kQV.4UicpCnz8hyeefBD4JHKAlZToL2K0EN1HV.u6sXpP1Xy')`)
	return err
}

func migrate1to2MySQL(d *DatabaseAccessor) error {
	// Move device blacklist to blacklist table
	bd, err := d.DB.Query(`SELECT "mac" FROM "device" WHERE "blacklisted" = 1`)
	if err != nil {
		return err
	}
	defer bd.Close()

	rowCount := 0
	sql := `INSERT INTO "blacklist" ("value") VALUES `

	for bd.Next() {
		var mac string
		if err := bd.Scan(&mac); err != nil {
			return err
		}
		sql += "('" + mac + "'), "
		rowCount++
	}

	if rowCount == 0 {
		return nil
	}

	sql = sql[:len(sql)-2]
	if _, err := d.DB.Exec(sql); err != nil {
		return err
	}
	return nil
}
