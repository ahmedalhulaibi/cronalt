package main

import (
	"context"
	"fmt"
	"sync"
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

	ctx := context.Background()

	js := job.NewStore()

	scheduler, _ := cronalt.NewScheduler(10, cronalt.WithJobStore(js), cronalt.WithLogger(loggylog))
	dscheduler := &DynamicJobScheduler{
		ctx:       ctx,
		state:     stopped,
		scheduler: scheduler,
		jobStore:  js,
	}

	dscheduler.Schedule(cronexpr.MustParse("0 * * * * * *"), foo{})
	dscheduler.Schedule(cronalt.Every(5*time.Second), panicker{})

	go func() {
		var flip bool
		ticker := time.NewTicker(time.Second * 10)
		flipjob := echoJob{}
		for range ticker.C {
			if flip {
				dscheduler.Schedule(cronalt.Every(time.Second), flipjob)
			} else {
				dscheduler.Remove(flipjob)
			}

			flip = !flip
		}
	}()

	dscheduler.Wait(ctx)
}

type jobStore interface {
	Remove(name string) error
}

type schedulerState byte

const (
	stopped schedulerState = iota
	started
)

type DynamicJobScheduler struct {
	ctx        context.Context
	cancelFunc context.CancelFunc
	scheduler  *cronalt.Scheduler
	jobStore   jobStore
	state      schedulerState
	locker     sync.Mutex
}

func (d *DynamicJobScheduler) Schedule(jt job.Timer, j job.Job) error {
	d.locker.Lock()
	defer d.locker.Unlock()

	err := d.scheduler.Schedule(jt, j)
	if err != nil {
		return err
	}

	d.restart()

	return nil
}

func (d *DynamicJobScheduler) Remove(j job.Job) error {
	d.locker.Lock()
	defer d.locker.Unlock()

	err := d.jobStore.Remove(j.Name())
	if err != nil {
		return err
	}

	d.restart()

	return nil
}

func (d *DynamicJobScheduler) restart() {
	if d.state == started {
		d.cancelFunc()
		d.state = stopped
	}

	ctx, cancelFunc := context.WithCancel(d.ctx)
	d.cancelFunc = cancelFunc

	go d.scheduler.Start(ctx)
	d.state = started
}

func (d *DynamicJobScheduler) Wait(ctx context.Context) {
	<-ctx.Done()
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
