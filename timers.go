package cronalt

import (
	"time"

	"github.com/ahmedalhulaibi/cronalt/job"
	"github.com/gorhill/cronexpr"
)

// jobTimer acts as a scheduler for a specific job
// given the prev start time of the job and current time
// the jobTimer implementation should return duration to wait until running again
type jobTimer interface {
	job.Timer
}

type durationTimer struct {
	time.Duration
}

var _ jobTimer = (*durationTimer)(nil)

func (d durationTimer) Next(prevStart time.Time) time.Time {
	return prevStart.Add(d.Duration)
}

func Every(d time.Duration) durationTimer {
	return durationTimer{d}
}

var _ jobTimer = (*cronexpr.Expression)(nil)
