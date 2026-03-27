package services

import (
	"context"
	"log/slog"
	"os"
	"sync"
)

// ──────────────────────────────────────────────────────────
// ADVANCED FEATURE: Concurrent Worker Pool
// Demonstrates: Fan-out/fan-in, channels, context, WaitGroup
// ──────────────────────────────────────────────────────────

var wpLogger = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
	Level: slog.LevelInfo,
}))

// Job represents a unit of work to be processed by a worker.
type Job struct {
	ID      int
	Name    string
	Execute func() (interface{}, error)
}

// Result holds the output of a completed job.
type Result struct {
	JobID  int
	Name   string
	Output interface{}
	Err    error
}

// WorkerPool manages a pool of goroutine workers.
type WorkerPool struct {
	workerCount int
	jobs        chan Job
	results     chan Result
	ctx         context.Context
	cancel      context.CancelFunc
	wg          sync.WaitGroup
}

// NewWorkerPool creates a worker pool with the given number of workers.
// Uses context.Context for graceful shutdown.
func NewWorkerPool(workerCount, jobBufferSize int) *WorkerPool {
	ctx, cancel := context.WithCancel(context.Background())
	pool := &WorkerPool{
		workerCount: workerCount,
		jobs:        make(chan Job, jobBufferSize),
		results:     make(chan Result, jobBufferSize),
		ctx:         ctx,
		cancel:      cancel,
	}
	return pool
}

// Start launches all workers as goroutines.
// Each worker reads from the jobs channel and writes to results.
func (wp *WorkerPool) Start() {
	for i := 0; i < wp.workerCount; i++ {
		wp.wg.Add(1)
		go wp.worker(i)
	}

	// Close results channel when all workers are done.
	go func() {
		wp.wg.Wait()
		close(wp.results)
	}()

	wpLogger.Info("Worker pool started",
		slog.Int("workers", wp.workerCount),
	)
}

// worker is the goroutine that processes jobs.
func (wp *WorkerPool) worker(id int) {
	defer wp.wg.Done()

	for {
		select {
		case <-wp.ctx.Done():
			wpLogger.Info("Worker shutting down", slog.Int("worker_id", id))
			return
		case job, ok := <-wp.jobs:
			if !ok {
				return // Channel closed
			}

			wpLogger.Info("Worker processing job",
				slog.Int("worker_id", id),
				slog.Int("job_id", job.ID),
				slog.String("job_name", job.Name),
			)

			output, err := job.Execute()
			wp.results <- Result{
				JobID:  job.ID,
				Name:   job.Name,
				Output: output,
				Err:    err,
			}
		}
	}
}

// Submit adds a job to the pool.
func (wp *WorkerPool) Submit(job Job) {
	wp.jobs <- job
}

// Close signals no more jobs and waits for completion.
func (wp *WorkerPool) Close() {
	close(wp.jobs)
}

// Shutdown cancels all workers immediately.
func (wp *WorkerPool) Shutdown() {
	wp.cancel()
	close(wp.jobs)
}

// Results returns the results channel to read from.
func (wp *WorkerPool) Results() <-chan Result {
	return wp.results
}

// ProcessBatch submits multiple jobs and collects all results.
// Convenience method that demonstrates the full fan-out/fan-in pattern.
func ProcessBatch(workerCount int, jobs []Job) []Result {
	pool := NewWorkerPool(workerCount, len(jobs))
	pool.Start()

	// Submit all jobs (fan-out)
	go func() {
		for _, job := range jobs {
			pool.Submit(job)
		}
		pool.Close()
	}()

	// Collect all results (fan-in)
	var results []Result
	for result := range pool.Results() {
		if result.Err != nil {
			wpLogger.Error("Job failed",
				slog.Int("job_id", result.JobID),
				slog.String("error", result.Err.Error()),
			)
		} else {
			wpLogger.Info("Job completed",
				slog.Int("job_id", result.JobID),
				slog.String("job_name", result.Name),
			)
		}
		results = append(results, result)
	}

	return results
}


