// This source file is part of the Packet Guardian project.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//
// +build dbmysql dball

package db

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/go-sql-driver/mysql" // MySQL driver
	"github.com/lfkeitel/verbose"
	"github.com/packet-guardian/packet-guardian/src/common"
)

func init() {
	RegisterDatabaseAccessor("mysql", newmySQLDBInit())
}

type mySQLDB struct {
	createFuncs  map[string]func(*common.DatabaseAccessor) error
	migrateFuncs []func(*common.DatabaseAccessor) error
}

func newmySQLDBInit() *mySQLDB {
	m := &mySQLDB{}

	m.createFuncs = map[string]func(*common.DatabaseAccessor) error{
		"blacklist":     m.createBlacklistTable,
		"device":        m.createDeviceTable,
		"lease":         m.createLeaseTable,
		"lease_history": m.createLeaseHistoryTable,
		"settings":      m.createSettingTable,
		"user":          m.createUserTable,
	}

	m.migrateFuncs = []func(*common.DatabaseAccessor) error{
		1: m.migrate1,
	}

	return m
}

func (m *mySQLDB) connect(d *common.DatabaseAccessor, c *common.Config) error {
	if c.Database.Port == 0 {
		c.Database.Port = 3306
	}

	mc := &mysql.Config{
		User:   c.Database.Username,
		Passwd: c.Database.Password,
		Net:    "tcp",
		Addr:   fmt.Sprintf("%s:%d", c.Database.Address, c.Database.Port),
		DBName: c.Database.Name,
		Strict: true,
	}
	var err error
	d.DB, err = sql.Open("mysql", mc.FormatDSN())
	if err != nil {
		return err
	}

	if err := d.DB.Ping(); err != nil {
		return err
	}

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
	return nil
}

func (m *mySQLDB) createTables(d *common.DatabaseAccessor) error {
	rows, err := d.DB.Query(`SHOW TABLES`)
	if err != nil {
		return err
	}
	defer rows.Close()

	tables := make(map[string]bool)
	for _, table := range common.DatabaseTableNames {
		tables[table] = false
	}

	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err != nil {
			return err
		}
		tables[tableName] = true
	}

	for table, create := range m.createFuncs {
		if !tables[table] {
			if err := create(d); err != nil {
				return err
			}
		}
	}
	return nil
}

func (m *mySQLDB) migrateTables(d *common.DatabaseAccessor) error {
	var currDBVer int
	verRow := d.DB.QueryRow(`SELECT "value" FROM "settings" WHERE "id" = 'db_version'`)
	if verRow == nil {
		return errors.New("Failed to get database version")
	}
	verRow.Scan(&currDBVer)

	common.SystemLogger.WithFields(verbose.Fields{
		"current-version": currDBVer,
		"active-version":  DBVersion,
	}).Debug("Database Versions")

	// No migration needed
	if currDBVer == DBVersion {
		return nil
	}

	if currDBVer > DBVersion {
		return errors.New("Database is too new, can't rollback")
	}

	neededMigrations := m.migrateFuncs[currDBVer:DBVersion]
	for _, migrate := range neededMigrations {
		if migrate == nil {
			continue
		}
		if err := migrate(d); err != nil {
			return err
		}
	}

	_, err := d.DB.Exec(`UPDATE "settings" SET "value" = ? WHERE "id" = 'db_version'`, DBVersion)
	return err
}

func (m *mySQLDB) init(d *common.DatabaseAccessor, c *common.Config) error {
	if err := m.connect(d, c); err != nil {
		return err
	}
	d.Driver = "mysql"

	if err := m.createTables(d); err != nil {
		return err
	}

	return m.migrateTables(d)
}

func (m *mySQLDB) createBlacklistTable(d *common.DatabaseAccessor) error {
	sql := `CREATE TABLE "blacklist" (
	    "id" INTEGER PRIMARY KEY AUTO_INCREMENT NOT NULL,
	    "value" VARCHAR(255) NOT NULL UNIQUE KEY,
	    "comment" TEXT
	) ENGINE=InnoDB DEFAULT CHARSET=utf8 AUTO_INCREMENT=1`

	_, err := d.DB.Exec(sql)
	return err
}

func (m *mySQLDB) createDeviceTable(d *common.DatabaseAccessor) error {
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

func (m *mySQLDB) createLeaseTable(d *common.DatabaseAccessor) error {
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

func (m *mySQLDB) createLeaseHistoryTable(d *common.DatabaseAccessor) error {
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

func (m *mySQLDB) createSettingTable(d *common.DatabaseAccessor) error {
	sql := `CREATE TABLE "settings" (
	    "id" VARCHAR(255) PRIMARY KEY NOT NULL,
	    "value" TEXT
	) ENGINE=InnoDB DEFAULT CHARSET=utf8`

	if _, err := d.DB.Exec(sql); err != nil {
		return err
	}

	_, err := d.DB.Exec(`INSERT INTO "settings" ("id", "value") VALUES ('db_version', ?)`, DBVersion)
	return err
}

func (m *mySQLDB) createUserTable(d *common.DatabaseAccessor) error {
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

func (m *mySQLDB) migrate1(d *common.DatabaseAccessor) error {
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
