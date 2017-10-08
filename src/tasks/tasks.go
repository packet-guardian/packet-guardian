// This source file is part of the Packet Guardian project.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package tasks

import (
	"errors"
	"time"

	"github.com/lfkeitel/verbose"
	"github.com/packet-guardian/packet-guardian/src/common"
)

type Job func(e *common.Environment) (string, error)

var jobs = make(map[string]Job)

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
		e.Log.WithFields(verbose.Fields{
			"package": "tasks:old-devices",
			"default": "1h",
		}).Notice("Invalid JobSchedulerWakeUp setting, using default")
		d = time.Hour
	}
	for {
		e.Log.WithFields(verbose.Fields{
			"package":  "tasks",
			"duration": d.String(),
		}).Info("Job scheduler sleeping")
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
		e.Log.WithFields(verbose.Fields{
			"package": "tasks",
			"job":     name,
		}).Info("Running scheduled job")
		result, err := job(e)
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
