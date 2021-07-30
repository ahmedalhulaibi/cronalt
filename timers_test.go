package cronalt

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func Test_durationTimer_Next(t *testing.T) {
	nowFixture := time.Date(2021, 01, 01, 01, 01, 01, 01, time.UTC)

	type fields struct {
		Duration time.Duration
	}
	type args struct {
		prevStart time.Time
	}
	tests := map[string]struct {
		fields fields
		args   args
		want   time.Time
	}{
		"Should return 1 second into the future": {
			fields: fields{
				Duration: time.Second,
			},
			args: args{
				prevStart: nowFixture,
			},
			want: nowFixture.Add(time.Second),
		},
	}
	for name, tt := range tests {
		tt := tt
		name := name
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			d := durationTimer{
				Duration: tt.fields.Duration,
			}
			assert.Equal(t, tt.want, d.Next(tt.args.prevStart))
		})
	}
}
