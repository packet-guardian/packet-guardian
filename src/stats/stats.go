// This source file is part of the Packet Guardian project.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package stats

import (
	"strings"
	"time"

	"github.com/onesimus-systems/packet-guardian/src/common"
)

type LeaseStats map[string]*NetworkStats

type NetworkStats struct {
	Registered   int
	Unregistered int
}

// GetLeaseStats returns a map of NetworkStats to network names
func GetLeaseStats(e *common.Environment) LeaseStats {
	sql := `SELECT "network", "registered", COUNT(*) FROM "lease" WHERE "end" > ? GROUP BY "network", "registered";`
	rows, err := e.DB.Query(sql, time.Now().Unix())
	if err != nil {
		e.Log.WithField("ErrMsg", err).Error("SQL statement failed")
		return nil
	}
	defer rows.Close()
	counts := make(LeaseStats)
	for rows.Next() {
		var network string
		var registered bool
		var count int

		if err := rows.Scan(&network, &registered, &count); err != nil {
			continue
		}

		network = strings.Title(network)

		if _, ok := counts[network]; !ok {
			counts[network] = &NetworkStats{}
		}

		if registered {
			counts[network].Registered = count
		} else {
			counts[network].Unregistered = count
		}
	}
	return counts
}

// GetDeviceStats return the total count and average per user
func GetDeviceStats(e *common.Environment) (int, int) {
	row := e.DB.QueryRow(`SELECT COUNT(*), COUNT(DISTINCT "username") FROM "device"`)
	total := 0
	distinct := 0
	if err := row.Scan(&total, &distinct); err != nil {
		e.Log.WithField("ErrMsg", err).Error("SQL statement failed")
		return 0, 0
	}
	if distinct == 0 {
		distinct = 1
	}
	return total, (total / distinct)
}
