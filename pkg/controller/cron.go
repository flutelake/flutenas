package controller

import (
	"context"
	"errors"
	"sync"

	"github.com/robfig/cron/v3"
)

type CronJob struct {
	cron *cron.Cron
	jobs map[string]cron.EntryID
	mu   sync.Mutex
}

// type Job struct {
// 	UniqueName   string
// 	ScheduleRule string
// 	JobFunc      func()
// }

func NewCronJob() *CronJob {
	return &CronJob{
		cron: cron.New(cron.WithSeconds()),
		jobs: make(map[string]cron.EntryID),
	}
}

func (c *CronJob) Start() {
	c.cron.Start()
}

func (c *CronJob) Stop() context.Context {
	return c.cron.Stop()
}

func (c *CronJob) AddJob(name, schedule string, job func()) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if _, exists := c.jobs[name]; exists {
		return ErrJobExists
	}

	id, err := c.cron.AddFunc(schedule, job)
	if err != nil {
		return err
	}

	c.jobs[name] = id
	return nil
}

func (c *CronJob) RemoveJob(name string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	id, exists := c.jobs[name]
	if !exists {
		return ErrJobNotFound
	}

	c.cron.Remove(id)
	delete(c.jobs, name)
	return nil
}

func (c *CronJob) ListJobs() []string {
	c.mu.Lock()
	defer c.mu.Unlock()

	names := make([]string, 0, len(c.jobs))
	for name := range c.jobs {
		names = append(names, name)
	}
	return names
}

var (
	ErrJobExists   = errors.New("job already exists")
	ErrJobNotFound = errors.New("job not found")
)
