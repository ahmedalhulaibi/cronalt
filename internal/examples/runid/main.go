package main

import (
	"context"
	"time"

	"github.com/ahmedalhulaibi/loggy"
	"go.uber.org/zap"

	"github.com/ahmedalhulaibi/cronalt"
	cronaltcontext "github.com/ahmedalhulaibi/cronalt/extensions/context"
	cronaltrunid "github.com/ahmedalhulaibi/cronalt/extensions/runid"
	"github.com/ahmedalhulaibi/cronalt/job"
)

func main() {
	logger, _ := zap.NewProduction(zap.AddCallerSkip(2))

	loggylog := localLogger{loggy.New(logger.Sugar()).WithFields("scheduler", cronaltrunid.RunIDContextKey)}
	defer loggylog.Sync()

	ejob := echoJob{
		logger: loggylog,
	}

	scheduler, _ := cronalt.NewScheduler(10, cronalt.WithLogger(loggylog))
	scheduler.Schedule(
		cronalt.Every(5*time.Second),
		job.Decorate(
			ejob,
			cronaltcontext.WithContext(cronaltrunid.RunIDBuilder(ejob, cronaltrunid.UUIDGenerator)),
		),
	)

	scheduler.Start(context.Background())
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
