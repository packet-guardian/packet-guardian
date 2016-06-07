// This source file is part of the Packet Guardian project.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package tasks

import (
	"fmt"
	"time"

	"github.com/onesimus-systems/packet-guardian/src/common"
)

func init() {
	RegisterJob("Purge old users", cleanUpExpiredUsers)
}

// Deletes users that expired 7 days ago
func cleanUpExpiredUsers(e *common.Environment) (string, error) {
	now := time.Now()
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
		e.Log.Infof("Purging user %s", username)
		i++
	}

	if i == 0 {
		return "No users to purge", nil
	}

	sql := `DELETE FROM "user" WHERE "valid_forever" = 0 AND "valid_end" < ?`
	results, err := e.DB.Exec(sql, now.Unix())
	if err != nil {
		return "", err
	}
	numOfRows, _ := results.RowsAffected()
	return fmt.Sprintf("Purged %d users", numOfRows), nil
}
