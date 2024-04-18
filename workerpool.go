package ansync

import (
	"sync"
)

type (
	workerPool struct {
		maxWorker   int
		action      func(interface{}) (interface{}, error)
		queuedTaskC chan Task
		wg          sync.WaitGroup

		onClose func()
		handler func(onSuccess func(resp <-chan interface{}), onFailed func(errs <-chan error))
	}

	Options func(*workerPool)

	WorkerPool interface {
		// Await wait all task to finish
		Await()

		// Handle response and error from finished task
		Handle(onSuccess func(resp <-chan interface{}), onFailed func(errs <-chan error))

		// Submit add another task
		Submit(Task)
	}
)

func (wp *workerPool) addTask(task Task) {
	wp.queuedTaskC <- task
}

func (wp *workerPool) run(outChan chan<- interface{}, errChan chan<- error) {
	for i := 0; i < wp.maxWorker; i++ {
		wID := i + 1

		go func(workerID int) {
			for req := range wp.queuedTaskC {
				ret, err := req()
				if err != nil {
					errChan <- err
					wp.wg.Done()
					continue
				}
				outChan <- ret
				wp.wg.Done()
			}
		}(wID)
	}
}

func (wp *workerPool) close(onClose func()) {
	close(wp.queuedTaskC)
	if onClose != nil {
		onClose()
	}
}

func defaultWorkerPool() *workerPool {
	return &workerPool{
		maxWorker: 1,
	}
}

func WithMaxWorker(maxWorker int) Options {
	return func(pool *workerPool) {
		if maxWorker < 1 {
			return
		}
		pool.maxWorker = maxWorker
	}
}

func (wp *workerPool) Handle(onSuccess func(resp <-chan interface{}), onFailed func(errs <-chan error)) {
	wp.handler(func(resp <-chan interface{}) {
		onSuccess(resp)
	}, func(errs <-chan error) {
		onFailed(errs)
	})
}

func (wp *workerPool) Await() {
	wp.onClose()
}

func (wp *workerPool) Submit(task Task) {
	wp.wg.Add(1)
	go func() {
		wp.addTask(task)
	}()
}

func DoWithWorkerPool(tasks []Task, opts ...Options) WorkerPool {
	wp := defaultWorkerPool()
	for _, opt := range opts {
		opt(wp)
	}
	wp.queuedTaskC = make(chan Task, wp.maxWorker)

	var (
		done = make(chan interface{}, 1)
		fail = make(chan error, 1)
	)

	wp.run(done, fail)
	wp.wg.Add(len(tasks))

	go func() {
		for _, task := range tasks {
			wp.addTask(task)
		}
	}()

	wp.onClose = func() {
		wp.wg.Wait()
		wp.close(func() {
			close(done)
			close(fail)
		})
	}

	wp.handler = func(onSuccess func(resp <-chan interface{}), onFailed func(errs <-chan error)) {
		go func() {
			onSuccess(done)
		}()
		go func() {
			onFailed(fail)
		}()
	}

	return wp
}
