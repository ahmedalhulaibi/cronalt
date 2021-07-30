package cronalt

import "time"

type timeProvider struct{}

func (timeProvider) Now() time.Time {
	return time.Now()
}
