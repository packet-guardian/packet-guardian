// This source file is part of the Packet Guardian project.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package tasks

import (
	"errors"
	"time"

	"github.com/lfkeitel/verbose/v4"
	"github.com/packet-guardian/packet-guardian/src/common"
	"github.com/packet-guardian/packet-guardian/src/models/stores"
)

type Job func(*common.Environment, stores.StoreCollection) (string, error)

var jobs = make(map[string]Job)

func RegisterJob(name string, job Job) error {
	if _, exists := jobs[name]; exists {
		return errors.New("Job already exists")
	}
	jobs[name] = job
	return nil
}

func StartTaskScheduler(e *common.Environment, stores stores.StoreCollection) {
	d, err := time.ParseDuration(e.Config.Core.JobSchedulerWakeUp)
	if err != nil {
		e.Log.WithFields(verbose.Fields{
			"package": "tasks:old-devices",
			"default": "1h",
		}).Notice("Invalid JobSchedulerWakeUp setting, using default")
		d = time.Hour
	}

	go flaggedDevicesTask(e, stores)
	for {
		e.Log.WithFields(verbose.Fields{
			"package":  "tasks",
			"duration": d.String(),
		}).Info("Job scheduler sleeping")
		time.Sleep(d)
		runJobs(e, stores)
	}
}

func runJobs(e *common.Environment, stores stores.StoreCollection) {
	defer func() {
		if r := recover(); r != nil {
			e.Log.WithField("Err", r).
				Alert("Recovered from panic running scheduled jobs")
		}
	}()

	// Run through a list of tasks
	for name, job := range jobs {
		e.Log.WithFields(verbose.Fields{
			"package": "tasks",
			"job":     name,
		}).Info("Running scheduled job")
		result, err := job(e, stores)
		if err != nil {
			e.Log.WithFields(verbose.Fields{
				"package": "tasks",
				"job":     name,
				"error":   err,
			}).Error("Job failed")
			continue
		}
		e.Log.WithFields(verbose.Fields{
			"package": "tasks",
			"job":     name,
			"result":  result,
		}).Info("Job finished")
	}
}
