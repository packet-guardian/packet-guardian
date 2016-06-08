// This source file is part of the Packet Guardian project.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package tasks

import (
	"fmt"
	"time"

	"github.com/usi-lfkeitel/packet-guardian/src/common"
)

func init() {
	RegisterJob("Purge old devices", cleanUpOldDevices)
}

// Deletes devices that haven't been seen in the last 6 months
// and devices which expired 7 days ago
func cleanUpOldDevices(e *common.Environment) (string, error) {
	// Use a constant date
	now := time.Now()
	d, err := time.ParseDuration(e.Config.Registration.RollingExpirationLength)
	if err != nil {
		e.Log.Error("Invalid RollingExpirationLength setting. Defaulting to 4380h")
		d = time.Duration(4380) * time.Hour
	}
	d = -d // This is how long ago a device was last seen, must be negative

	sqlSel := `SELECT "mac" FROM "device" WHERE "expires" != 0 AND ("last_seen" < ? OR ("expires" != 1 AND "expires" < ?))`
	rows, err := e.DB.Query(sqlSel, now.Add(d).Unix(), now.Unix())
	if err != nil {
		return "", err
	}
	defer rows.Close()

	i := 0
	for rows.Next() {
		var mac string
		rows.Scan(&mac)
		e.Log.Infof("Purging device %s", mac)
		i++
	}

	if i == 0 {
		return "No devices to purge", nil
	}

	sql := `DELETE FROM "device" WHERE "expires" != 0 AND ("last_seen" < ? OR ("expires" != 1 AND "expires" < ?))`
	results, err := e.DB.Exec(sql, now.Add(d).Unix(), now.Unix())
	if err != nil {
		return "", err
	}
	numOfRows, _ := results.RowsAffected()
	return fmt.Sprintf("Purged %d devices", numOfRows), nil
}
