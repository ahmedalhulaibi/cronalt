package main

import (
	"context"
	"fmt"
	"time"

	"github.com/ahmedalhulaibi/loggy"
	"go.uber.org/zap"

	"github.com/ahmedalhulaibi/cronalt"
	cronaltcircuitbreaker "github.com/ahmedalhulaibi/cronalt/extensions/circuitbreaker"
	cronaltcontext "github.com/ahmedalhulaibi/cronalt/extensions/context"
	cronaltrunid "github.com/ahmedalhulaibi/cronalt/extensions/runid"
	"github.com/ahmedalhulaibi/cronalt/job"
)

type Breaker struct {
	threshold        uint
	errorOccurrences uint
}

func (b *Breaker) BreakerFunc() func(error) bool {
	return func(err error) bool {
		if err == nil {
			return false
		}

		b.errorOccurrences++
		return b.errorOccurrences >= b.threshold
	}
}

func main() {
	logger, _ := zap.NewProduction(zap.AddCallerSkip(2))

	loggylog := localLogger{loggy.New(logger.Sugar()).WithFields("scheduler", cronaltrunid.RunIDContextKey)}
	defer loggylog.Sync()

	ejob := echoJob{
		logger: loggylog,
	}

	errorThrowerJob := errorThrower{}

	circuitBreaker := Breaker{
		threshold: 2,
	}

	ctx, cancelFunc := context.WithCancel(context.Background())

	scheduler, _ := cronalt.NewScheduler(10, cronalt.WithLogger(loggylog))
	scheduler.Schedule(
		cronalt.Every(5*time.Second),
		job.Decorate(
			ejob,
			cronaltcontext.WithContext(cronaltrunid.RunIDBuilder(ejob, cronaltrunid.UUIDGenerator)),
		),
	)
	scheduler.Schedule(
		cronalt.Every(10*time.Second),
		job.Decorate(
			errorThrowerJob,
			cronaltcontext.WithContext(cronaltrunid.RunIDBuilder(errorThrowerJob, cronaltrunid.UUIDGenerator)),
			cronaltcircuitbreaker.WithCircuitBreaker(circuitBreaker.BreakerFunc(), cancelFunc),
		),
	)

	scheduler.Start(ctx)
}

type localLogger struct {
	loggy.Logger
}

func (l localLogger) getArgs(args ...cronalt.KeyVal) []loggy.KeyVal {
	argsL := make([]loggy.KeyVal, 0, len(args))
	for _, arg := range args {
		argsL = append(argsL, loggy.KeyVal(arg))
	}
	return argsL
}

func (l localLogger) Info(ctx context.Context, msg string, args ...cronalt.KeyVal) {
	l.Logger.Infow(ctx, msg, l.getArgs(args...)...)
}

func (l localLogger) Error(ctx context.Context, msg string, args ...cronalt.KeyVal) {
	l.Logger.Errorw(ctx, msg, l.getArgs(args...)...)
}

func (l localLogger) Warn(ctx context.Context, msg string, args ...cronalt.KeyVal) {
	l.Logger.Warnw(ctx, msg, l.getArgs(args...)...)
}

type echoJob struct {
	logger localLogger
}

func (echoJob) Name() string {
	return "echoJob"
}

func (e echoJob) Runner() job.JobFn {
	return func(ctx context.Context) error {
		e.logger.Info(ctx, "echoJob did a thing")
		return nil
	}
}

type errorThrower struct{}

func (errorThrower) Name() string {
	return "errorThrower"
}

func (errorThrower) Runner() job.JobFn {
	return func(ctx context.Context) error {
		return fmt.Errorf("errorThrower failed")
	}
}
