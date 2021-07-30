package cronaltcircuitbreaker

import (
	"context"

	"github.com/ahmedalhulaibi/cronalt/job"
)

type circuitBreaker struct {
	breakerFunc breakerFunc
	cancelFunc  func()
	job         job.Job
}

type breakerFunc func(error) bool

func WithCircuitBreaker(bf breakerFunc, cancelFunc func()) job.Decorator {
	return func(j job.Job) job.Job {
		// This must be passed as a pointer since it maintains state
		return &circuitBreaker{
			breakerFunc: bf,
			cancelFunc:  cancelFunc,
			job:         j,
		}
	}
}

func (c *circuitBreaker) Name() string {
	return c.job.Name()
}

func (c *circuitBreaker) Runner() job.JobFn {
	return func(ctx context.Context) error {
		err := c.job.Runner()(ctx)
		if c.breakerFunc(err) {
			go c.cancelFunc()
		}

		return err
	}
}
