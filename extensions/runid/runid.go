package cronaltrunid

import (
	"context"

	"github.com/google/uuid"

	cronaltcontext "github.com/ahmedalhulaibi/cronalt/extensions/context"
	"github.com/ahmedalhulaibi/cronalt/job"
)

type RunIDGenerator func(j job.Job) string

func UUIDGenerator(j job.Job) string {
	return j.Name() + "-" + uuid.NewString()
}

const RunIDContextKey string = "run_id"

func RunIDBuilder(j job.Job, r RunIDGenerator) cronaltcontext.ContextBuilder {
	return func(ctx context.Context) context.Context {
		return context.WithValue(ctx, RunIDContextKey, r(j))
	}
}
