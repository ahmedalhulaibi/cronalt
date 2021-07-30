package main

import (
	"context"
	"fmt"
	"time"

	"github.com/ahmedalhulaibi/cronalt/job"
	"github.com/ahmedalhulaibi/loggy"
	"github.com/gorhill/cronexpr"
	"go.uber.org/zap"

	"github.com/ahmedalhulaibi/cronalt"
)

func main() {
	logger, _ := zap.NewProduction(zap.AddCallerSkip(2))

	loggylog := localLogger{loggy.New(logger.Sugar())}
	defer loggylog.Sync()

	scheduler, _ := cronalt.NewScheduler(10, cronalt.WithLogger(loggylog))

	scheduler.Schedule(cronalt.Every(time.Second), echoJob{})
	scheduler.Schedule(cronexpr.MustParse("0 * * * * * *"), foo{})
	scheduler.Schedule(cronalt.Every(5*time.Second), panicker{})

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

func echo(_ context.Context) error {
	fmt.Println("hello")
	return nil
}

type echoJob struct{}

func (echoJob) Name() string {
	return "echoJob"
}

func (echoJob) Runner() job.JobFn {
	return echo
}

type foo struct{}

func (foo) Name() string {
	return "foo"
}

func (foo) Runner() job.JobFn {
	return func(ctx context.Context) error {
		fmt.Println("foo")
		return nil
	}
}

type panicker struct{}

func (panicker) Name() string {
	return "panicker"
}

func (panicker) Runner() job.JobFn {
	return func(ctx context.Context) error {
		panic(fmt.Errorf("panicker panicking"))
	}
}
