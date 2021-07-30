package cronalt

import "context"

type KeyVal struct {
	Key string
	Val interface{}
}

type logger interface {
	Info(ctx context.Context, msg string, args ...KeyVal)
	Error(ctx context.Context, msg string, args ...KeyVal)
	Warn(ctx context.Context, msg string, args ...KeyVal)
}

type noopLogger struct{}

func (n noopLogger) Info(_ context.Context, _ string, _ ...KeyVal) {}

func (n noopLogger) Warn(_ context.Context, _ string, _ ...KeyVal) {}

func (n noopLogger) Error(_ context.Context, _ string, _ ...KeyVal) {}
