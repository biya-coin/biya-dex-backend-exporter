package collectors

import (
	"context"
	"errors"
	"log/slog"
	"sync"
	"sync/atomic"
	"time"

	"github.com/biya-coin/biya-dex-backend-exporter/internal/metrics"
)

type Collector interface {
	Run(ctx context.Context) error
}

type Job struct {
	Name     string
	Interval time.Duration
	Collector
}

func NewJob(name string, interval time.Duration, c Collector) Job {
	return Job{Name: name, Interval: interval, Collector: c}
}

type Scheduler struct {
	log   *slog.Logger
	m     *metrics.Metrics
	jobs  []Job
	ready atomic.Bool

	mu           sync.Mutex
	jobReadyOnce map[string]bool
}

func NewScheduler(log *slog.Logger, m *metrics.Metrics, jobs []Job) *Scheduler {
	s := &Scheduler{
		log:          log,
		m:            m,
		jobs:         jobs,
		jobReadyOnce: make(map[string]bool, len(jobs)),
	}
	return s
}

func (s *Scheduler) Ready() bool {
	return s.ready.Load()
}

func (s *Scheduler) Run(ctx context.Context) error {
	if len(s.jobs) == 0 {
		return errors.New("no jobs configured")
	}

	var wg sync.WaitGroup
	for _, job := range s.jobs {
		j := job
		if j.Interval <= 0 {
			s.log.Warn("skip job with non-positive interval", "job", j.Name, "interval", j.Interval)
			continue
		}
		wg.Add(1)
		go func() {
			defer wg.Done()
			s.runJobLoop(ctx, j)
		}()
	}

	<-ctx.Done()
	wg.Wait()
	return nil
}

func (s *Scheduler) runJobLoop(ctx context.Context, job Job) {
	// 首次立即跑一次，避免 exporter 启动后长时间无数据
	s.runOnce(ctx, job)

	t := time.NewTicker(job.Interval)
	defer t.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			s.runOnce(ctx, job)
		}
	}
}

func (s *Scheduler) runOnce(ctx context.Context, job Job) {
	start := time.Now()
	err := job.Collector.Run(ctx)
	dur := time.Since(start).Seconds()

	s.m.ObserveDuration(job.Name, dur)
	if err != nil {
		s.m.SetGauge("biya_exporter_scrape_success", map[string]string{"source": job.Name}, 0)
		s.log.Error("collector run failed", "collector", job.Name, "duration_s", dur, "err", err)
		return
	}

	s.m.SetGauge("biya_exporter_scrape_success", map[string]string{"source": job.Name}, 1)
	s.log.Debug("collector run ok", "collector", job.Name, "duration_s", dur)

	s.markJobReady(job.Name)
}

func (s *Scheduler) markJobReady(jobName string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.jobReadyOnce[jobName] {
		return
	}
	s.jobReadyOnce[jobName] = true

	// 全部 job 首次成功后置 ready
	for _, j := range s.jobs {
		if !s.jobReadyOnce[j.Name] {
			return
		}
	}
	s.ready.Store(true)
}
