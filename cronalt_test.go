package cronalt

import (
	"bufio"
	"bytes"
	"context"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/ahmedalhulaibi/cronalt/job"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func Test_timeUntilNextRun(t *testing.T) {
	nowFixture := time.Date(2021, 01, 01, 01, 01, 01, 01, time.UTC)

	type args struct {
		nextExpectedRun time.Time
		now             time.Time
	}
	tests := map[string]struct {
		args args
		want time.Duration
	}{
		"Should return 0 seconds when next expected run is now": {
			args: args{
				nextExpectedRun: nowFixture,
				now:             nowFixture,
			},
			want: 0 * time.Second,
		},
		"Should return 0 seconds when nextExpectedRun was in the past": {
			args: args{
				nextExpectedRun: nowFixture.Add(-1 * time.Second),
				now:             nowFixture,
			},
			want: 0 * time.Second,
		},
		"Should return 1 second when nextExpectedRun is 1 second after now": {
			args: args{
				nextExpectedRun: nowFixture.Add(1 * time.Second),
				now:             nowFixture,
			},
			want: time.Second,
		},
		"Should return 100 seconds when nextExpectedRun is 100 seconds after now": {
			args: args{
				nextExpectedRun: nowFixture.Add(100 * time.Second),
				now:             nowFixture,
			},
			want: 100 * time.Second,
		},
	}
	for name, tt := range tests {
		tt := tt
		name := name
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, timeUntilNextRun(tt.args.nextExpectedRun, tt.args.now))
		})
	}
}

type JobFn = job.JobFn

type mockJob struct {
	mock.Mock
}

func (m *mockJob) Name() string {
	args := m.Called()

	return args.String(0)
}

func (m *mockJob) Runner() JobFn {
	args := m.Called()

	return args.Get(0).(JobFn)
}

type mockLogger struct {
	mock.Mock
}

func (m *mockLogger) Info(ctx context.Context, msg string, args ...KeyVal) {
	m.Called(ctx, msg, args)
}

func (m *mockLogger) Error(ctx context.Context, msg string, args ...KeyVal) {
	m.Called(ctx, msg, args)
}

func (m *mockLogger) Warn(ctx context.Context, msg string, args ...KeyVal) {
	m.Called(ctx, msg, args)
}

func Test_recoverJob(t *testing.T) {
	ctxFixture := context.Background()
	type args struct {
		ctx context.Context
		job job.Job
		log logger
	}
	tests := map[string]struct {
		args      args
		expectErr bool
	}{
		// TODO: Add test cases.
		"Should recover with error": {
			args: args{
				ctx: ctxFixture,
			},
			expectErr: true,
		},
		"Should recover with no error": {
			args: args{
				ctx: ctxFixture,
			},
			expectErr: false,
		},
	}
	for name, tt := range tests {
		tt := tt
		name := name
		t.Run(name, func(t *testing.T) {
			{
				job := &mockJob{}

				job.On("Name").Return("jobname")
				defer job.AssertExpectations(t)

				tt.args.job = job
			}

			{
				logger := &mockLogger{}

				if tt.expectErr {
					logger.On("Error", ctxFixture, "cronalt.Scheduler recovered", []KeyVal{
						{"job", "jobname"},
						{"error", "error message"},
					})
				} else {
					logger.On("Error", ctxFixture, "cronalt.Scheduler recovered", []KeyVal{
						{"job", "jobname"},
					})
				}

				defer logger.AssertExpectations(t)

				tt.args.log = logger
			}

			func() {
				defer recoverJob(tt.args.ctx, tt.args.job, tt.args.log)

				if tt.expectErr {
					panic(fmt.Errorf("error message"))
				}

				panic("")
			}()

		})
	}
}

type fmtLogger struct {
	w io.Writer
}

func (fl fmtLogger) Info(ctx context.Context, msg string, keyVals ...KeyVal) {
	fmt.Fprintf(fl.w, getFormat("INFO", len(keyVals)), getArgs(msg, keyVals...)...)
}

func (fl fmtLogger) Error(ctx context.Context, msg string, keyVals ...KeyVal) {
	fmt.Fprintf(fl.w, getFormat("ERROR", len(keyVals)), getArgs(msg, keyVals...)...)
}

func (fl fmtLogger) Warn(ctx context.Context, msg string, keyVals ...KeyVal) {
	fmt.Fprintf(fl.w, getFormat("WARN", len(keyVals)), getArgs(msg, keyVals...)...)
}

func getFormat(level string, lenKeyVals int) string {
	format := level + " %s "
	pair := "%v:%v"
	end := "\n"

	for i := 0; i < lenKeyVals; i++ {
		format = format + pair
	}

	return format + end
}

func getArgs(msg string, kv ...KeyVal) []interface{} {
	args := make([]interface{}, 0, (len(kv)*2)+1)
	args = append(args, msg)
	args = append(args, keyValsToInterface(kv...)...)
	return args
}

func keyValsToInterface(keyVals ...KeyVal) []interface{} {
	args := make([]interface{}, 0, len(keyVals)*2)

	for _, kv := range keyVals {
		args = append(args, kv.Key, kv.Val)
	}

	return args
}

var update = flag.Bool("update", false, "update .golden files")

type mockScheduler struct {
	called   bool
	mockTime time.Time
}

func (m *mockScheduler) Next(prev time.Time) time.Time {
	if !m.called {
		m.called = true
		return m.mockTime
	}

	return m.mockTime.Add(time.Hour * 24 * 365 * 100)
}

type cancellerJob struct {
	cancel func()
	err    error
}

func (c cancellerJob) Name() string {
	return "canceller"
}

func (c cancellerJob) Runner() JobFn {
	return func(ctx context.Context) error {
		go c.cancel()
		return c.err
	}
}

func TestScheduler_run(t *testing.T) {
	// TODO: Golden file is not necessary in this test, just use a mock logger + assertions
	nowFixture := time.Date(2021, 01, 01, 01, 01, 01, 01, time.UTC)
	t.Run("Should run without errors", func(t *testing.T) {
		// setup log writer
		var b bytes.Buffer
		bw := bufio.NewWriter(&b)

		ctx, cancel := context.WithCancel(context.Background())

		s, err := NewScheduler(1, WithLogger(fmtLogger{w: bw}))
		require.NoError(t, err)

		s.Schedule(&mockScheduler{mockTime: nowFixture}, cancellerJob{cancel: cancel})
		s.Start(ctx)

		// Flush logs
		require.NoError(t, bw.Flush())

		gp := filepath.Join("internal/test_fixtures", t.Name()+".golden")
		if *update {
			t.Log("updating golden file")
			require.NoError(t, os.MkdirAll(filepath.Dir(gp), 0744))
			require.NoError(t, ioutil.WriteFile(gp, b.Bytes(), 0744))
		}

		g, err := ioutil.ReadFile(gp)
		require.NoError(t, err)

		require.Equal(t, b.Bytes(), g)
	})
	t.Run("Should run with error", func(t *testing.T) {
		// setup log writer
		var b bytes.Buffer
		bw := bufio.NewWriter(&b)

		ctx, cancel := context.WithCancel(context.Background())

		s, err := NewScheduler(1, WithLogger(fmtLogger{w: bw}))
		require.NoError(t, err)

		s.Schedule(&mockScheduler{mockTime: nowFixture}, cancellerJob{cancel: cancel, err: fmt.Errorf("custom_err")})
		s.Start(ctx)

		// Flush logs
		require.NoError(t, bw.Flush())

		gp := filepath.Join("internal/test_fixtures", t.Name()+".golden")
		if *update {
			t.Log("updating golden file")
			require.NoError(t, os.MkdirAll(filepath.Dir(gp), 0744))
			require.NoError(t, ioutil.WriteFile(gp, b.Bytes(), 0744))
		}

		g, err := ioutil.ReadFile(gp)
		require.NoError(t, err)

		require.Equal(t, b.Bytes(), g)
	})
}

func TestScheduler_Schedule(t *testing.T) {
	t.Run("Should return ErrJobExists when the job is already scheduled", func(t *testing.T) {
		j := &mockJob{}

		j.On("Name").Return("jobname")
		defer j.AssertExpectations(t)

		s, err := NewScheduler(10)
		require.NoError(t, err)

		{
			err := s.Schedule(Every(time.Second), j)
			require.NoError(t, err)
		}

		{
			err := s.Schedule(Every(time.Second), j)
			require.EqualError(t, err, fmt.Errorf("%w:%s", job.ErrJobExists, "jobname").Error())
		}
	})
}

func TestNewScheduler(t *testing.T) {
	tests := map[string]struct {
		maxConcurrentJobs int
		want              error
	}{
		"Should return error when maxConcurrentJobs is less than zero": {
			maxConcurrentJobs: -1,
			want:              ErrMaxConcurrentJobsZero,
		},
		"Should return error when maxConcurrentJobs is equal to zero": {
			maxConcurrentJobs: 0,
			want:              ErrMaxConcurrentJobsZero,
		},
		"Should not return any errors when maxConcurrentJobs is greater than zero": {
			maxConcurrentJobs: 1,
			want:              nil,
		},
	}
	for name, tt := range tests {
		tt := tt
		t.Run(name, func(t *testing.T) {
			_, err := NewScheduler(tt.maxConcurrentJobs)
			if tt.want != nil {
				require.EqualError(t, err, tt.want.Error())
			} else {
				require.NoError(t, err)
			}
		})
	}
}
