package workerpool

import (
	"context"
	"sync"
	"sync/atomic"
)

const (
	defaultSize = 10
)

type WorkerPool interface {
	Start()
	Stop()
	Do(tasks ...Task)
	DoAsync(tasks ...Task)
}

type workerPool struct {
	idx     int32
	workers []Worker
	size    int32
}

type WorkerPoolOptions struct {
	workerFactory WorkerFactory
}

type WorkerPoolOption func(*WorkerPoolOptions)

func WithWorkerFactory(workerFactory WorkerFactory) WorkerPoolOption {
	return func(o *WorkerPoolOptions) {
		o.workerFactory = workerFactory
	}
}

var defaultWorkerPoolOptions = WorkerPoolOptions{
	workerFactory: NewWorker,
}

func New(size int, opts ...WorkerPoolOption) WorkerPool {
	defaultOptions := defaultWorkerPoolOptions
	for _, opt := range opts {
		opt(&defaultOptions)
	}
	if size <= 0 {
		size = defaultSize
	}
	wp := &workerPool{
		size: int32(size),
	}
	for i := size; i > 0; i-- {
		worker := defaultOptions.workerFactory()
		wp.workers = append(wp.workers, worker)
	}
	return wp
}

func (wp *workerPool) Start() {
	for _, worker := range wp.workers {
		worker.Start()
	}
}

func (wp *workerPool) Stop() {
	for _, worker := range wp.workers {
		worker.Stop()
	}
}

func (wp *workerPool) Do(tasks ...Task) {
	if len(tasks) == 0 {
		return
	}
	wg := &sync.WaitGroup{}
	wg.Add(len(tasks))
	for _, task := range tasks {
		idx := wp.nextIdx()
		wrappedTask := func(ctx context.Context) {
			defer wg.Done()
			task(ctx)
		}
		wp.workers[idx].AddTask(wrappedTask)
	}
	wg.Wait()
}

func (wp *workerPool) DoAsync(tasks ...Task) {
	if len(tasks) == 0 {
		return
	}
	for _, task := range tasks {
		idx := wp.nextIdx()
		wp.workers[idx].AddTask(task)
	}
}

func (wp *workerPool) nextIdx() int32 {
	idx := atomic.AddInt32(&wp.idx, 1)
	return idx % wp.size
}
