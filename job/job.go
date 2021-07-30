package job

import (
	"context"
	"time"
)

type Config interface {
	Timer() Timer
	Job() Job
}

// Timer acts as a scheduler for a specific job
// given the prev start time of the job and current time
// the Timer implementation should return duration to wait until running again
type Timer interface {
	Next(prevStart time.Time) time.Time
}

type Job interface {
	Name() string
	Runner() JobFn
}

type JobFn func(ctx context.Context) error

type Decorator func(Job) Job

func Decorate(j Job, decorators ...Decorator) Job {
	for _, d := range decorators {
		j = d(j)
	}

	return j
}
