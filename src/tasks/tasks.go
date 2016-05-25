// This source file is part of the Packet Guardian project.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package tasks

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/onesimus-systems/packet-guardian/src/common"
)

type Job func(e *common.Environment) (string, error)

var jobs = make(map[string]Job)

func init() {
	RegisterJob("Purge old devices", cleanUpOldDevices)
	RegisterJob("Purge old users", cleanUpExpiredUsers)
}

func RegisterJob(name string, job Job) error {
	if _, exists := jobs[name]; exists {
		return errors.New("Job already exists")
	}
	jobs[name] = job
	return nil
}

func StartTaskScheduler(e *common.Environment) {
	d, err := time.ParseDuration(e.Config.Core.JobSchedulerWakeUp)
	if err != nil {
		e.Log.Error("Invalid JobSchedulerWakeUp setting. Defaulting to 1h")
		d = time.Hour
	}
	for {
		e.Log.Infof("Job scheduler sleeping for %s", d.String())
		time.Sleep(d)
		runJobs(e)
	}
}

func runJobs(e *common.Environment) {
	defer func() {
		if r := recover(); r != nil {
			e.Log.WithField("Err", r).
				Alert("Recovered from panic running scheduled jobs")
		}
	}()

	// Run through a list of tasks
	for name, job := range jobs {
		e.Log.Infof("Running scheduled job '%s'", name)
		result, err := job(e)
		if err != nil {
			e.Log.WithField("Err", err).Errorf("'%s' job failed", name)
			continue
		}
		e.Log.Infof("'%s' job finished - %s", name, result)
	}
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

	sqlSel := `SELECT "mac" FROM "device" WHERE "expires" != 0 AND ("last_seen" < ? OR "expires" < ?)`
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

	sql := `DELETE FROM "device" WHERE "expires" != 0 AND ("last_seen" < ? OR "expires" < ?)`
	results, err := e.DB.Exec(sql, now.Add(d).Unix(), now.Unix())
	if err != nil {
		return "", err
	}
	numOfRows, _ := results.RowsAffected()
	return fmt.Sprintf("Purged %d devices", numOfRows), nil
}

// Deletes users that expired 7 days ago
func cleanUpExpiredUsers(e *common.Environment) (string, error) {
	now := time.Now()
	sqlSel := `SELECT "username" FROM "user" WHERE "valid_forever" = 0 AND "valid_end" < ?`
	rows, err := e.DB.Query(sqlSel, now.Unix())
	if err == sql.ErrNoRows {
		return "No users purged", nil
	} else if err != nil {
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
