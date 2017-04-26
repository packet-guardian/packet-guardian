// This source file is part of the Packet Guardian project.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package tasks

import (
	"fmt"
	"time"

	"github.com/packet-guardian/packet-guardian/src/common"
)

func init() {
	RegisterJob("Purge old users", cleanUpExpiredUsers)
}

// Deletes users that expired 7 days ago
func cleanUpExpiredUsers(e *common.Environment) (string, error) {
	now := time.Now().Add(time.Duration(-7) * 24 * time.Hour)
	sqlSel := `SELECT "username" FROM "user" WHERE "valid_forever" = 0 AND "valid_end" < ?`
	rows, err := e.DB.Query(sqlSel, now.Unix())
	if err != nil {
		return "", err
	}
	defer rows.Close()

	i := 0
	for rows.Next() {
		var username string
		rows.Scan(&username)
		e.Log.WithField("username", username).Info("TASK - Deleting user")
		i++
	}

	if i == 0 {
		return "No users to delete", nil
	}

	sql := `DELETE FROM "user" WHERE "valid_forever" = 0 AND "valid_end" < ?`
	results, err := e.DB.Exec(sql, now.Unix())
	if err != nil {
		return "", err
	}
	numOfRows, _ := results.RowsAffected()
	return fmt.Sprintf("Deleted %d users", numOfRows), nil
}
