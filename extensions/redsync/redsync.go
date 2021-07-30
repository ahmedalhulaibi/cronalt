package cronaltredsync

import (
	"context"

	"github.com/go-redsync/redsync/v4"

	"github.com/ahmedalhulaibi/cronalt/job"
)

type locker struct {
	rsmutex *redsync.Mutex
	job     job.Job
}

func WithLocker(rs *redsync.Redsync, opts ...redsync.Option) job.Decorator {
	return func(j job.Job) job.Job {
		return &locker{
			job:     j,
			rsmutex: rs.NewMutex(j.Name(), opts...),
		}
	}
}

func (l locker) Name() string {
	return l.job.Name()
}

func (l locker) Runner() job.JobFn {
	return func(ctx context.Context) error {
		err := l.rsmutex.Lock()
		if err != nil {
			return err
		}

		defer func() { l.rsmutex.Unlock() }()
		return l.job.Runner()(ctx)
	}
}
