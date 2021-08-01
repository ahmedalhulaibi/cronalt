package cronaltcounter

import (
	"context"
	"sync/atomic"

	"github.com/ahmedalhulaibi/cronalt/job"
)

type JobCounter struct {
	job   job.Job
	count uint32
}

func New() *JobCounter {
	return &JobCounter{}
}

func WithJobCounter(jc *JobCounter) job.Decorator {
	return func(j job.Job) job.Job {
		jc.job = j
		return jc
	}
}

func (jc *JobCounter) Name() string {
	return jc.job.Name()
}

func (jc *JobCounter) Runner() job.JobFn {
	return func(ctx context.Context) error {
		atomic.AddUint32(&jc.count, 1)
		return jc.job.Runner()(ctx)
	}
}

func (jc *JobCounter) Count() uint32 {
	return atomic.LoadUint32(&jc.count)
}
