# cronalt

Golang job scheduling
- Runs in-process
- Type-safe (avoids reflection)
- Pluggable and extensible

cronalt is inspired by [gocron](https://github.com/go-co-op/gocron)

Install using go modules

```shell
go get github.com/ahmedalhulaibi/cronalt@v1.0.0
```

## Getting started

See [`internal/examples/basic`](internal/examples/basic) directory for a simple example

## Concepts

### Job

The job is simply aware of the task it's provided. It is defined as an interface. Jobs can be extended/wrapped using a decorator to extend functionality.

### Job Timer

A Job Timer informs the Scheduler how long to wait before triggering the next run for a given Job. The Job Timer is also defined as an interface to allow for irregular scheduling.

### Scheduler

The Scheduler orchestrates all Jobs. It starts all the jobs and stops all the jobs. For each job, an individual goroutine is kicked off with its Job Timer informing the routine how the job should be scheduled.

## How Do I...?

### How do I lock a job in a distributed system (K8s, Nomad, etc.)?

Decorate your job with a locker implementation.

See [`internal/examples/redsync`](internal/examples/redsync) for an example.

### How do I handle an error if my job fails?

Decorate your job with an error handler implementation.

See [`internal/examples/errorhandler`](internal/examples/errorhandler) for an example.

### How do I stop the scheduler if a job keeps failing?

Decorate all your jobs with a circuit breaker which cancels the parent context. This will stop the entire process, not just the individual routine.

See [`internal/examples/circuitbreaker`](internal/examples/circuitbreaker) for an example

### How do I propagate custom fields in context?

Decorate your job with a context decorator.

See [`internal/examples/runid`](internal/examples/runid) for an example

### How do I dynamically add and remove jobs?