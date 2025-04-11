// This source file is part of the Packet Guardian project.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//
//go:build dbmysql || dball
// +build dbmysql dball

package db

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/packet-guardian/packet-guardian/src/models/stores"

	"github.com/go-sql-driver/mysql" // MySQL driver
	"github.com/lfkeitel/verbose/v4"
	"github.com/packet-guardian/packet-guardian/src/common"
)

func init() {
	RegisterDatabaseAccessor("mysql", newmySQLDBInit())
}

type mySQLDB struct {
	createFuncs  map[string]func(*common.DatabaseAccessor) error
	migrateFuncs []migrateFunc
}

func newmySQLDBInit() *mySQLDB {
	m := &mySQLDB{}

	m.createFuncs = map[string]func(*common.DatabaseAccessor) error{
		"blacklist":        m.createBlacklistTable,
		"device":           m.createDeviceTable,
		"lease":            m.createLeaseTable,
		"settings":         m.createSettingTable,
		"user":             m.createUserTable,
		"account_delegate": m.createDelegateTable,
	}

	m.migrateFuncs = []migrateFunc{
		1: m.migrateFrom1,
		2: m.migrateFrom2,
		3: m.migrateFrom3,
		4: m.migrateFrom4,
		5: m.migrateFrom5,
	}

	return m
}

func (m *mySQLDB) connect(d *common.DatabaseAccessor, c *common.Config) error {
	if c.Database.Port == 0 {
		c.Database.Port = 3306
	}

	mc := mysql.NewConfig()
	mc.User = c.Database.Username
	mc.Passwd = c.Database.Password
	mc.Net = "tcp"
	mc.Addr = fmt.Sprintf("%s:%d", c.Database.Address, c.Database.Port)
	mc.DBName = c.Database.Name

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

	if !strings.Contains(mode, "ANSI") {
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
			fmt.Printf("Creating table %s\n", table)
			if err := create(d); err != nil {
				return err
			}
		}
	}
	return nil
}

func (m *mySQLDB) migrateTables(d *common.DatabaseAccessor, c *common.Config) error {
	var currDBVer int
	verRow := d.DB.QueryRow(`SELECT "value" FROM "settings" WHERE "id" = 'db_version'`)
	if verRow == nil {
		return errors.New("Failed to get database version")
	}
	verRow.Scan(&currDBVer)

	// No migration needed
	if currDBVer == DBVersion {
		common.SystemLogger.WithFields(verbose.Fields{
			"current-database-version": currDBVer,
		}).Debug("Database schema is up-to-date")
		return nil
	}

	common.SystemLogger.WithFields(verbose.Fields{
		"current-database-version":     currDBVer,
		"application-database-version": DBVersion,
	}).Debug("Applying database migrations...")

	if currDBVer > DBVersion {
		return errors.New("Database is too new, can't rollback")
	}

	neededMigrations := m.migrateFuncs[currDBVer:DBVersion]
	for _, migrate := range neededMigrations {
		if migrate == nil {
			continue
		}
		if err := migrate(d, c); err != nil {
			return err
		}
	}

	_, err := d.DB.Exec(`UPDATE "settings" SET "value" = ? WHERE "id" = 'db_version'`, DBVersion)

	common.SystemLogger.WithFields(verbose.Fields{
		"current-database-version": DBVersion,
	}).Debug("Database migrations applied successfully")
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

	return m.migrateTables(d, c)
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
		"description" TEXT,
		"last_seen" INTEGER NOT NULL,
		"flagged" TINYINT DEFAULT 0,
		"notes" TEXT
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
		"valid_forever" TINYINT DEFAULT 1,
		"ui_group" VARCHAR(20) NOT NULL DEFAULT 'default',
		"api_group" VARCHAR(20) NOT NULL DEFAULT 'disable',
		"allow_status_api" TINYINT DEFAULT 0,
		"notes" TEXT
	) ENGINE=InnoDB DEFAULT CHARSET=utf8 AUTO_INCREMENT=4;`

	if _, err := d.DB.Exec(sql); err != nil {
		return err
	}

	_, err := d.DB.Exec(`INSERT INTO "user"
			("id", "username", "password", "ui_group") VALUES
			(1, 'admin', '$2a$10$rZfN/gdXZdGYyLtUb6LF.eHOraDes3ibBECmWic2I3SocMC0L2Lxa', 'admin'),
			(2, 'helpdesk', '$2a$10$ICCdq/OyZBBoNPTRmfgntOnujD6INGv7ZAtA/Xq6JIdRMO65xCuNC', 'helpdesk'),
			(3, 'readonly', '$2a$10$02NG6kQV.4UicpCnz8hyeefBD4JHKAlZToL2K0EN1HV.u6sXpP1Xy', 'readonly')`)
	return err
}

func (m *mySQLDB) createDelegateTable(d *common.DatabaseAccessor) error {
	sql := `CREATE TABLE "account_delegate" (
		"id" INTEGER PRIMARY KEY AUTO_INCREMENT NOT NULL,
		"user_id" INTEGER NOT NULL,
		"delegate" VARCHAR(255) NOT NULL,
		"permissions" CHAR(2) NOT NULL DEFAULT 0,
		CONSTRAINT user_delegates UNIQUE ("user_id", "delegate")
	) ENGINE=InnoDB DEFAULT CHARSET=utf8;`

	_, err := d.DB.Exec(sql)
	return err
}

func (m *mySQLDB) migrateFrom1(d *common.DatabaseAccessor, c *common.Config) error {
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

func (m *mySQLDB) migrateFrom2(d *common.DatabaseAccessor, c *common.Config) error {
	sql := `ALTER TABLE "user" ADD COLUMN (
		"ui_group" VARCHAR(20) NOT NULL DEFAULT 'default',
		"api_group" VARCHAR(20) NOT NULL DEFAULT 'disable',
		"allow_status_api" TINYINT DEFAULT 0
	);`

	if _, err := d.DB.Exec(sql); err != nil {
		return err
	}

	common.RegisterSystemInitFunc(migrateUserPermissions)
	return nil
}

func migrateUserPermissions(e *common.Environment) error {
	if err := migrateUserGroup(e, e.Config.Auth.AdminUsers, "ui", "admin"); err != nil {
		return err
	}
	if err := migrateUserGroup(e, e.Config.Auth.HelpDeskUsers, "ui", "helpdesk"); err != nil {
		return err
	}
	if err := migrateUserGroup(e, e.Config.Auth.ReadOnlyUsers, "ui", "readonly"); err != nil {
		return err
	}
	if err := migrateUserGroup(e, e.Config.Auth.APIReadOnlyUsers, "api", "readonly-api"); err != nil {
		return err
	}
	if err := migrateUserGroup(e, e.Config.Auth.APIReadWriteUsers, "api", "readwrite-api"); err != nil {
		return err
	}
	if err := migrateUserGroup(e, e.Config.Auth.APIStatusUsers, "api-status", ""); err != nil {
		return err
	}
	return nil
}

func migrateUserGroup(e *common.Environment, members []string, group, groupName string) error {
	// This usage of GetUserStore is an exception, getting dependencies injected this
	// far would be too much trouble for little benefit.
	users := stores.GetUserStore(e)

	for _, username := range members {
		user, err := users.GetUserByUsername(username)
		if err != nil {
			return err
		}

		switch group {
		case "ui":
			user.UIGroup = groupName
		case "api":
			user.APIGroup = groupName
		case "api-status":
			user.AllowStatusAPI = true
		}

		if err := user.Save(); err != nil {
			return err
		}
	}
	return nil
}

func (m *mySQLDB) migrateFrom3(d *common.DatabaseAccessor, c *common.Config) error {
	sql := `DROP TABLE IF EXISTS "lease_history"`
	_, err := d.DB.Exec(sql)
	return err
}

func (m *mySQLDB) migrateFrom4(d *common.DatabaseAccessor, c *common.Config) error {
	sql := `ALTER TABLE "device" ADD COLUMN (
		"flagged" TINYINT DEFAULT 0
	);`
	_, err := d.DB.Exec(sql)
	return err
}

func (m *mySQLDB) migrateFrom5(d *common.DatabaseAccessor, c *common.Config) error {
	sql := `ALTER TABLE "user" ADD COLUMN (
		"notes" TEXT
	);`
	if _, err := d.DB.Exec(sql); err != nil {
		return err
	}

	sql = `ALTER TABLE "device" ADD COLUMN (
		"notes" TEXT
	);`
	_, err := d.DB.Exec(sql)
	return err
}
