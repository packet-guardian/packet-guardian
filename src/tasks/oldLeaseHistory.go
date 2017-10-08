// This source file is part of the Packet Guardian project.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package tasks

import (
	"fmt"
	"time"

	"github.com/lfkeitel/verbose"
	"github.com/packet-guardian/packet-guardian/src/common"
)

func init() {
	RegisterJob("Purge old lease history", cleanUpOldLeaseHistory)
}

// Deletes lease history where "end" is before now() - duration
func cleanUpOldLeaseHistory(e *common.Environment) (string, error) {
	// Use a constant date
	now := time.Now()
	d, err := time.ParseDuration(e.Config.Leases.DeleteAfter)
	if err != nil {
		e.Log.WithFields(verbose.Fields{
			"package": "tasks:old-lease-history",
			"default": "672h",
		}).Notice("Invalid DeleteAfter setting, using default")
		d = 672 * time.Hour
	}
	d = -d

	sql := `DELETE FROM "lease_history" WHERE "end" < ?`
	results, err := e.DB.Exec(sql, now.Add(d).Unix())
	if err != nil {
		return "", err
	}
	numOfRows, _ := results.RowsAffected()
	return fmt.Sprintf("Deleted %d history entries", numOfRows), nil
}
