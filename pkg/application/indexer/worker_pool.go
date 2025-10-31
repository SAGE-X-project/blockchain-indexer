package indexer

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/sage-x-project/blockchain-indexer/pkg/infrastructure/logger"
	"go.uber.org/zap"
)

// WorkerPool manages a pool of workers for concurrent block processing
type WorkerPool struct {
	workerCount int
	jobQueue    chan Job
	results     chan Result
	wg          sync.WaitGroup
	ctx         context.Context
	cancel      context.CancelFunc
	logger      *logger.Logger

	// Metrics
	activeWorkers  int32
	totalJobs      uint64
	completedJobs  uint64
	failedJobs     uint64
	totalDuration  int64 // milliseconds
}

// Job represents a unit of work
type Job struct {
	ID      string
	Type    JobType
	Payload interface{}
}

// JobType represents the type of job
type JobType int

const (
	// JobTypeBlock represents a block processing job
	JobTypeBlock JobType = iota
	// JobTypeBlockRange represents a block range processing job
	JobTypeBlockRange
	// JobTypeTransaction represents a transaction processing job
	JobTypeTransaction
)

// String returns the string representation of JobType
func (j JobType) String() string {
	switch j {
	case JobTypeBlock:
		return "block"
	case JobTypeBlockRange:
		return "block_range"
	case JobTypeTransaction:
		return "transaction"
	default:
		return "unknown"
	}
}

// Result represents the result of a job
type Result struct {
	JobID     string
	Success   bool
	Error     error
	Duration  time.Duration
	Payload   interface{}
}

// JobHandler is a function that processes a job
type JobHandler func(ctx context.Context, job Job) Result

// WorkerPoolConfig holds worker pool configuration
type WorkerPoolConfig struct {
	WorkerCount int
	QueueSize   int
	ResultSize  int
}

// DefaultWorkerPoolConfig returns default worker pool configuration
func DefaultWorkerPoolConfig() *WorkerPoolConfig {
	return &WorkerPoolConfig{
		WorkerCount: 10,
		QueueSize:   100,
		ResultSize:  100,
	}
}

// NewWorkerPool creates a new worker pool
func NewWorkerPool(config *WorkerPoolConfig, handler JobHandler, logger *logger.Logger) *WorkerPool {
	if config == nil {
		config = DefaultWorkerPoolConfig()
	}

	ctx, cancel := context.WithCancel(context.Background())

	pool := &WorkerPool{
		workerCount: config.WorkerCount,
		jobQueue:    make(chan Job, config.QueueSize),
		results:     make(chan Result, config.ResultSize),
		ctx:         ctx,
		cancel:      cancel,
		logger:      logger,
	}

	// Start workers
	for i := 0; i < config.WorkerCount; i++ {
		pool.wg.Add(1)
		go pool.worker(i, handler)
	}

	pool.logger.Info("worker pool started",
		zap.Int("workers", config.WorkerCount),
		zap.Int("queue_size", config.QueueSize),
	)

	return pool
}

// worker is the worker goroutine
func (p *WorkerPool) worker(id int, handler JobHandler) {
	defer p.wg.Done()

	atomic.AddInt32(&p.activeWorkers, 1)
	defer atomic.AddInt32(&p.activeWorkers, -1)

	p.logger.Debug("worker started", zap.Int("worker_id", id))

	for {
		select {
		case <-p.ctx.Done():
			p.logger.Debug("worker stopped", zap.Int("worker_id", id))
			return

		case job, ok := <-p.jobQueue:
			if !ok {
				p.logger.Debug("job queue closed", zap.Int("worker_id", id))
				return
			}

			p.logger.Debug("worker processing job",
				zap.Int("worker_id", id),
				zap.String("job_id", job.ID),
				zap.String("job_type", job.Type.String()),
			)

			startTime := time.Now()

			// Process job
			result := handler(p.ctx, job)
			result.JobID = job.ID
			result.Duration = time.Since(startTime)

			// Update metrics
			atomic.AddUint64(&p.completedJobs, 1)
			atomic.AddInt64(&p.totalDuration, result.Duration.Milliseconds())

			if !result.Success {
				atomic.AddUint64(&p.failedJobs, 1)
			}

			// Send result
			select {
			case p.results <- result:
			case <-p.ctx.Done():
				return
			default:
				// Result channel full, log warning
				p.logger.Warn("result channel full, dropping result",
					zap.String("job_id", job.ID),
				)
			}
		}
	}
}

// Submit submits a job to the worker pool
func (p *WorkerPool) Submit(job Job) error {
	atomic.AddUint64(&p.totalJobs, 1)

	select {
	case <-p.ctx.Done():
		return fmt.Errorf("worker pool stopped")
	case p.jobQueue <- job:
		return nil
	default:
		return fmt.Errorf("job queue full")
	}
}

// SubmitWithTimeout submits a job with a timeout
func (p *WorkerPool) SubmitWithTimeout(job Job, timeout time.Duration) error {
	atomic.AddUint64(&p.totalJobs, 1)

	timer := time.NewTimer(timeout)
	defer timer.Stop()

	select {
	case <-p.ctx.Done():
		return fmt.Errorf("worker pool stopped")
	case p.jobQueue <- job:
		return nil
	case <-timer.C:
		return fmt.Errorf("submit timeout after %v", timeout)
	}
}

// Results returns the results channel
func (p *WorkerPool) Results() <-chan Result {
	return p.results
}

// Stop stops the worker pool
func (p *WorkerPool) Stop() {
	p.logger.Info("stopping worker pool")

	// Cancel context to stop workers
	p.cancel()

	// Close job queue
	close(p.jobQueue)

	// Wait for all workers to finish
	p.wg.Wait()

	// Close results channel
	close(p.results)

	p.logger.Info("worker pool stopped",
		zap.Uint64("total_jobs", atomic.LoadUint64(&p.totalJobs)),
		zap.Uint64("completed_jobs", atomic.LoadUint64(&p.completedJobs)),
		zap.Uint64("failed_jobs", atomic.LoadUint64(&p.failedJobs)),
	)
}

// Wait waits for all jobs to complete with a timeout
func (p *WorkerPool) Wait(timeout time.Duration) error {
	done := make(chan struct{})

	go func() {
		p.wg.Wait()
		close(done)
	}()

	timer := time.NewTimer(timeout)
	defer timer.Stop()

	select {
	case <-done:
		return nil
	case <-timer.C:
		return fmt.Errorf("wait timeout after %v", timeout)
	}
}

// Stats returns worker pool statistics
func (p *WorkerPool) Stats() *WorkerPoolStats {
	completed := atomic.LoadUint64(&p.completedJobs)
	totalDuration := atomic.LoadInt64(&p.totalDuration)

	avgDuration := int64(0)
	if completed > 0 {
		avgDuration = totalDuration / int64(completed)
	}

	return &WorkerPoolStats{
		WorkerCount:     p.workerCount,
		ActiveWorkers:   int(atomic.LoadInt32(&p.activeWorkers)),
		TotalJobs:       atomic.LoadUint64(&p.totalJobs),
		CompletedJobs:   completed,
		FailedJobs:      atomic.LoadUint64(&p.failedJobs),
		QueueLength:     len(p.jobQueue),
		QueueCapacity:   cap(p.jobQueue),
		ResultsLength:   len(p.results),
		ResultsCapacity: cap(p.results),
		AvgDuration:     time.Duration(avgDuration) * time.Millisecond,
	}
}

// WorkerPoolStats represents worker pool statistics
type WorkerPoolStats struct {
	WorkerCount     int
	ActiveWorkers   int
	TotalJobs       uint64
	CompletedJobs   uint64
	FailedJobs      uint64
	QueueLength     int
	QueueCapacity   int
	ResultsLength   int
	ResultsCapacity int
	AvgDuration     time.Duration
}

// String returns a string representation of the stats
func (s *WorkerPoolStats) String() string {
	return fmt.Sprintf(
		"Workers: %d/%d, Jobs: %d/%d, Failed: %d, Queue: %d/%d, Results: %d/%d, Avg: %v",
		s.ActiveWorkers,
		s.WorkerCount,
		s.CompletedJobs,
		s.TotalJobs,
		s.FailedJobs,
		s.QueueLength,
		s.QueueCapacity,
		s.ResultsLength,
		s.ResultsCapacity,
		s.AvgDuration,
	)
}

// IsHealthy checks if the worker pool is healthy
func (p *WorkerPool) IsHealthy() bool {
	stats := p.Stats()

	// Check if workers are active
	if stats.ActiveWorkers == 0 {
		return false
	}

	// Check if queue is not full
	if stats.QueueLength >= stats.QueueCapacity {
		return false
	}

	// Check if results channel is not full
	if stats.ResultsLength >= stats.ResultsCapacity {
		return false
	}

	return true
}
