// This source file is part of the Packet Guardian project.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//
// +build dbpostgres dball

package common

import (
	"errors"
)

func init() {
	dbInits["postgres"] = func(d *DatabaseAccessor, c *Config) error {
		return errors.New("PostgreSQL not implemented yet")
	}
}
