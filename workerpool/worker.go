package workerpool

import (
	"context"
	"sync"
	"sync/atomic"

	"github.com/wwq-2020/go.common/app"
)

type WorkerFactory func() Worker

type Task func(ctx context.Context)

type Worker interface {
	AddTask(Task)
	Start()
	Stop()
}

type worker struct {
	m      *sync.Mutex
	c      *sync.Cond
	tasks  []Task
	ctx    context.Context
	cancel context.CancelFunc
	done   int32
}

func NewWorker() Worker {
	m := &sync.Mutex{}
	c := sync.NewCond(m)
	ctx, cancel := context.WithCancel(context.TODO())
	return &worker{
		m:      m,
		c:      c,
		ctx:    ctx,
		cancel: cancel,
	}
}

func (w *worker) AddTask(task Task) {
	w.m.Lock()
	defer w.m.Unlock()
	w.tasks = append(w.tasks, task)
	w.c.Signal()

}

func (w *worker) Start() {
	app.GoAsyncForever(func() {
		for {
			if atomic.LoadInt32(&w.done) == 1 {
				return
			}
			w.eachLoop()
		}
	})
}

func (w *worker) Stop() {
	atomic.StoreInt32(&w.done, 1)
	w.cancel()
	w.m.Lock()
	defer w.m.Unlock()
	w.c.Signal()
}

func (w *worker) eachLoop() {
	task := w.fetchTask()
	if task == nil {
		return
	}
	task(w.ctx)
}

func (w *worker) fetchTask() Task {
	w.m.Lock()
	defer w.m.Unlock()
	if len(w.tasks) == 0 {
		w.c.Wait()
	}
	if len(w.tasks) == 0 {
		return nil
	}
	task := w.tasks[0]
	w.tasks = w.tasks[1:]
	return task
}
