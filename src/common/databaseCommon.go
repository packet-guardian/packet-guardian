// This source file is part of the Packet Guardian project.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package common

import (
	"database/sql"
	"errors"
)

const dbVersion = 2

type databaseInit func(*DatabaseAccessor, *Config) error

var dbInits = make(map[string]databaseInit)

type DatabaseAccessor struct {
	*sql.DB
	Driver string
}

func NewDatabaseAccessor(config *Config) (*DatabaseAccessor, error) {
	da := &DatabaseAccessor{}
	if f, ok := dbInits[config.Database.Type]; ok {
		return da, f(da, config)
	}
	return nil, errors.New("Database " + config.Database.Type + " not supported")
}
