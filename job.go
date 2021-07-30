package cronalt

import "github.com/ahmedalhulaibi/cronalt/job"

type jobStore interface {
	Add(j job.Config) error
	Remove(name string) error
	Get(name string) (job.Config, error)
	GetAll() []job.Config
}

type jobCfg struct {
	timer job.Timer
	j     job.Job
}

func (j jobCfg) Job() job.Job {
	return j.j
}

func (j jobCfg) Timer() job.Timer {
	return j.timer
}
