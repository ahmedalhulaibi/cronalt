package cronalterrorhandler

import (
	"context"

	"github.com/ahmedalhulaibi/cronalt/job"
)

type errorHandleFunc func(error) error

type errorHandler struct {
	job     job.Job
	handler errorHandleFunc
}

func WithErrorHandler(er errorHandleFunc) job.Decorator {
	return func(j job.Job) job.Job {
		return errorHandler{
			job:     j,
			handler: er,
		}
	}
}

func (er errorHandler) Name() string {
	return er.job.Name()
}

func (er errorHandler) Runner() job.JobFn {
	return func(ctx context.Context) error {
		err := er.job.Runner()(ctx)
		return er.handler(err)
	}
}
