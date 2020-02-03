package main

import (
	"errors"
	"sync"
	"time"
)

func main() {

}

type scheduled interface {
	nextRun() (time.Duration, error)
}

// Job defines a running job and allows to stop a scheduled job or run it.
type Job struct {
	id       string
	fn       func(id string)
	quit     chan bool
	err      error
	schedule scheduled
	sync.RWMutex
}

type recurrent struct {
	units  int
	period time.Duration
	done   bool
}

func (r *recurrent) nextRun() (time.Duration, error) {
	if r.units == 0 || r.period == 0 {
		return 0, errors.New("cannot set recurrent time with 0")
	}
	if !r.done {
		r.done = true
		return 0, nil
	}
	return time.Duration(r.units) * r.period, nil
}

// AddJob makes a job to run a func recurrently in a goroutine
func AddJob(id string, s scheduled, f func(id string)) {
	j := &Job{
		id:       id,
		schedule: s,
		fn:       f,
		quit:     make(chan bool),
	}
	j.run()
}

func (j *Job) run() error {
	next, err := j.schedule.nextRun()
	if err != nil {
		return err
	}
	go func(j *Job) {
		for {
			select {
			case <-j.quit:
				return
			case <-time.After(next):
				j.runJob()
			}
			next, _ = j.schedule.nextRun()
		}
	}(j)
	return nil
}

func (j *Job) runJob() {
	//call actual job
	j.fn(j.id)
	// call the database
	// if the db returns no object found err
	// quit the job
	// if the recurrent unit changed
	// update the schedule
}

// feedback checks the unassigned
func feedback() {
	for {

	}
}
