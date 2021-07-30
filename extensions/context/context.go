package cronaltcontext

import (
	"context"

	"github.com/ahmedalhulaibi/cronalt/job"
)

type ContextBuilder func(context.Context) context.Context

type contextDecorator struct {
	builder ContextBuilder
	job     job.Job
}

func WithContext(c ContextBuilder) job.Decorator {
	return func(j job.Job) job.Job {
		return contextDecorator{
			builder: c,
			job:     j,
		}
	}
}

func (c contextDecorator) Name() string {
	return c.job.Name()
}

func (c contextDecorator) Runner() job.JobFn {
	return func(ctx context.Context) error {
		return c.job.Runner()(c.builder(ctx))
	}
}
