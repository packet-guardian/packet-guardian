// This source file is part of the Packet Guardian project.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package tasks

import (
	"errors"
	"time"

	"github.com/onesimus-systems/packet-guardian/src/common"
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
