// This source file is part of the Packet Guardian project.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package common

import (
	"database/sql"
	"errors"
	"time"

	"github.com/lfkeitel/verbose"
)

const dbVersion = 2

type databaseInit func(*DatabaseAccessor, *Config) error

var dbInits = make(map[string]databaseInit)

type DatabaseAccessor struct {
	*sql.DB
	Driver string
}

func NewDatabaseAccessor(e *Environment) (*DatabaseAccessor, error) {
	da := &DatabaseAccessor{}
	if f, ok := dbInits[e.Config.Database.Type]; ok {
		var err error
		retries := 0
		dur, err := time.ParseDuration(e.Config.Database.RetryTimeout)
		if err != nil {
			return nil, errors.New("Invalid RetryTimeout")
		}
		e.Log.Debug(dur.String())

		// This loop will break when no error occurs when connecting to a database
		// Or when the number of attempted retries is greater than configured
		shutdownChan := e.SubscribeShutdown()

		for {
			err = f(da, e.Config)

			// If no error occured, break
			// If an error occured but retries is not set to inifinite and we've tried
			// too many times already, break
			if err == nil || (e.Config.Database.Retry != 0 && retries >= e.Config.Database.Retry) {
				break
			}

			retries++
			e.Log.WithFields(verbose.Fields{
				"Attempts":    retries,
				"MaxAttempts": e.Config.Database.Retry,
				"Timeout":     e.Config.Database.RetryTimeout,
				"Error":       err,
			}).Error("Failed to connect to database. Retrying after timeout.")

			select {
			case <-shutdownChan:
				return nil, err
			case <-time.After(dur):
			}
		}

		return da, err
	}
	return nil, errors.New("Database " + e.Config.Database.Type + " not supported")
}
