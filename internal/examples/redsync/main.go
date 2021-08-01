package main

import (
	"context"
	"fmt"
	"time"

	goredislib "github.com/go-redis/redis/v8"

	"github.com/ahmedalhulaibi/loggy"
	"github.com/go-redsync/redsync/v4"
	"github.com/go-redsync/redsync/v4/redis/goredis/v8"
	"go.uber.org/zap"

	"github.com/ahmedalhulaibi/cronalt"
	cronaltredsync "github.com/ahmedalhulaibi/cronalt/extensions/redsync"
	"github.com/ahmedalhulaibi/cronalt/job"
)

func main() {
	logger, _ := zap.NewProduction(zap.AddCallerSkip(2))

	loggylog := localLogger{loggy.New(logger.Sugar()).WithFields("scheduler")}
	defer loggylog.Sync()

	// Create a pool with go-redis (or redigo) which is the pool redisync will
	// use while communicating with Redis. This can also be any pool that
	// implements the `redis.Pool` interface.
	client := goredislib.NewClient(&goredislib.Options{
		Addr: "localhost:6379",
	})
	pool := goredis.NewPool(client) // or, pool := redigo.NewPool(...)

	// Create an instance of redisync to be used to obtain a mutual exclusion
	// lock.
	rs := redsync.New(pool)

	scheduler, _ := cronalt.NewScheduler(10, cronalt.WithLogger(loggylog))
	scheduler.Schedule(
		cronalt.Every(45*time.Second),
		job.Decorate(echoJob{}, cronaltredsync.WithLocker(rs, redsync.WithTries(1))),
	)

	scheduler2, _ := cronalt.NewScheduler(10, cronalt.WithLogger(loggylog))
	scheduler2.Schedule(
		cronalt.Every(45*time.Second),
		job.Decorate(echoJob{}, cronaltredsync.WithLocker(rs, redsync.WithTries(1))),
	)

	ctx := context.Background()
	go scheduler.Start(context.WithValue(ctx, "scheduler", "scheduler1"))
	scheduler2.Start(context.WithValue(ctx, "scheduler", "scheduler2"))
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
	time.Sleep(15 * time.Second)
	return nil
}

type echoJob struct{}

func (echoJob) Name() string {
	return "echoJob"
}

func (echoJob) Runner() job.JobFn {
	return echo
}
