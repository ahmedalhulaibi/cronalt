package job

import (
	"fmt"
	"sync"
)

type store struct {
	sync.RWMutex
	jobs map[string]Config
}

func NewStore() *store {
	return &store{
		jobs: make(map[string]Config),
	}
}

var (
	ErrJobExists       error = fmt.Errorf("job already scheduled with name")
	ErrJobDoesNotExist error = fmt.Errorf("job does not exist")
)

func (s *store) Add(j Config) error {
	s.Lock()
	defer s.Unlock()

	name := j.Job().Name()

	if _, ok := s.jobs[name]; ok {
		return fmt.Errorf("%w:%s", ErrJobExists, name)
	}

	// Key is the job name
	s.jobs[name] = j

	return nil
}

func (s *store) Remove(name string) error {
	s.Lock()
	defer s.Unlock()

	_, ok := s.jobs[name]
	if !ok {
		return ErrJobDoesNotExist
	}

	delete(s.jobs, name)

	return nil
}

func (s *store) Get(name string) (Config, error) {
	s.RLock()
	defer s.RUnlock()

	j, ok := s.jobs[name]
	if !ok {
		return j, ErrJobDoesNotExist
	}

	return j, nil
}

func (s *store) GetAll() []Config {
	s.RLock()
	defer s.RUnlock()

	copy := make([]Config, 0, len(s.jobs))

	for _, val := range s.jobs {
		copy = append(copy, val)
	}

	return copy
}
