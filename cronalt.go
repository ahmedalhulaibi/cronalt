package cronalt

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/ahmedalhulaibi/cronalt/job"
)

type empty struct{}

type clock interface {
	Now() time.Time
}

type waitGroup interface {
	Add(delta int)
	Wait()
	Done()
}

type Scheduler struct {
	// jobPool is a Semaphore to limit the number of concurrent jobs running at once
	jobPool chan empty
	jobs    jobStore
	wg      waitGroup
	log     logger
	clock   clock
}

var (
	ErrMaxConcurrentJobsZero error = fmt.Errorf("maxConcurrentJobs must be greater than zero")
)

func NewScheduler(maxConcurrentJobs int, opts ...SchedulerOption) (*Scheduler, error) {
	if maxConcurrentJobs <= 0 {
		return nil, ErrMaxConcurrentJobsZero
	}

	var wg sync.WaitGroup

	s := &Scheduler{
		jobPool: make(chan empty, maxConcurrentJobs),
		jobs:    job.NewStore(),
		log:     noopLogger{},
		clock:   timeProvider{},
		wg:      &wg,
	}

	for _, opt := range opts {
		s = opt(s)
	}

	return s, nil
}

type SchedulerOption func(s *Scheduler) *Scheduler

// WithLogger returns a SchedulerOption to inject a logger
func WithLogger(l logger) SchedulerOption {
	return func(s *Scheduler) *Scheduler {
		s.log = l
		return s
	}
}

// WithClock returns a SchedulerOption to inject a clock, default is time.Now
func WithClock(t clock) SchedulerOption {
	return func(s *Scheduler) *Scheduler {
		s.clock = t
		return s
	}
}

// WithWaitGroup returns a SchedulerOption to inject a waitgroup, default is sync.WaitGroup
// This can be used to inject your own instance of sync.WaitGroup
func WithWaitGroup(wg waitGroup) SchedulerOption {
	return func(s *Scheduler) *Scheduler {
		s.wg = wg
		return s
	}
}

// WithJobStore returns a SchedulerOption to inject a jobStore, default is job.store
func WithJobStore(js jobStore) SchedulerOption {
	return func(s *Scheduler) *Scheduler {
		s.jobs = js
		return s
	}
}

// Schedule registers a job and uses job function name as the job name
func (s *Scheduler) Schedule(jt job.Timer, j job.Job) error {
	return s.jobs.Add(jobCfg{timer: jt, j: j})
}

// Start starts all the scheduled jobs in their own go routine and blocks indefinitely or until context is cancelled
func (s *Scheduler) Start(ctx context.Context) {
	for _, pendingJob := range s.jobs.GetAll() {
		s.wg.Add(1)
		s.log.Info(ctx, "cronalt.Scheduler starting job", KeyVal{"job", pendingJob.Job().Name()})
		go s.run(ctx, pendingJob)
	}
	s.wg.Wait()
}

func (s *Scheduler) run(ctx context.Context, runJobCfg job.Config) {
	defer s.wg.Done()

	jobName := runJobCfg.Job().Name()

	timer := time.NewTimer(s.getTimeUntilNextRun(ctx, s.clock.Now(), runJobCfg))

	for {
		select {
		case <-ctx.Done():
			s.log.Info(ctx, "cronalt.Scheduler halted", KeyVal{"job", jobName})
			return
		case now := <-timer.C:
			s.log.Info(ctx, "cronalt.Scheduler queued", KeyVal{"job", jobName})

			// Acquire lock on job pool semaphore
			s.jobPool <- empty{}

			s.log.Info(ctx, "cronalt.Scheduler running", KeyVal{"job", jobName})

			if err := call(ctx, runJobCfg.Job(), s.log); err != nil {
				s.log.Error(
					ctx,
					"cronalt.Scheduler job completed with error",
					KeyVal{"job", jobName},
					KeyVal{"error", err.Error()},
				)
			}

			s.log.Info(ctx, "cronalt.Scheduler completed", KeyVal{"job", jobName})

			// Release lock on job pool semaphore
			<-s.jobPool

			// Use timer.Reset since we know the timer is expired and we can reset it
			timer.Reset(s.getTimeUntilNextRun(ctx, now, runJobCfg))
		}
	}
}

func (s *Scheduler) getTimeUntilNextRun(ctx context.Context, prevTime time.Time, runJobCfg job.Config) time.Duration {
	now := s.clock.Now()
	nextExpectedRun := runJobCfg.Timer().Next(prevTime)

	s.log.Info(
		ctx,
		"cronalt.Scheduler next run",
		KeyVal{"job", runJobCfg.Job().Name()},
		KeyVal{"next_run", nextExpectedRun.Format(time.RFC3339)},
	)

	return timeUntilNextRun(nextExpectedRun, now)
}

func timeUntilNextRun(nextExpectedRun, now time.Time) time.Duration {
	if nextExpectedRun.After(now) || nextExpectedRun.Equal(now) {
		return nextExpectedRun.Sub(now)
	}

	return 0
}

func call(ctx context.Context, job job.Job, log logger) error {
	defer recoverJob(ctx, job, log)

	return job.Runner()(ctx)
}

func recoverJob(ctx context.Context, job job.Job, log logger) {
	if r := recover(); r != nil {
		keys := make([]KeyVal, 1, 2)
		keys[0] = KeyVal{"job", job.Name()}

		if err, ok := r.(error); ok {
			keys = append(keys, KeyVal{"error", err.Error()})
		}

		log.Error(ctx, "cronalt.Scheduler recovered", keys...)
	}
}
