package ansync

import (
	"sync"
)

type (
	workerPool[T any] struct {
		maxWorker   int
		queuedTaskC chan Task[T]
		wg          sync.WaitGroup

		onClose func()
		handler func(onSuccess func(resp <-chan T), onFailed func(errs <-chan error))
	}

	WorkerPoolOption[T any] func(*workerPool[T])

	WorkerPool[T any] interface {
		// Await wait all task to finish
		Await()

		// Handle response and error from finished task
		Handle(onSuccess func(resp <-chan T), onFailed func(errs <-chan error))

		// Submit add another task
		Submit(Task[T])
	}
)

func (wp *workerPool[T]) addTask(task Task[T]) {
	wp.queuedTaskC <- task
}

func (wp *workerPool[T]) run(outChan chan<- T, errChan chan<- error) {
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

func (wp *workerPool[T]) close(onClose func()) {
	close(wp.queuedTaskC)
	if onClose != nil {
		onClose()
	}
}

func defaultWorkerPool[T any]() *workerPool[T] {
	return &workerPool[T]{
		maxWorker: 1,
	}
}

func WithMaxWorker[T any](maxWorker int) WorkerPoolOption[T] {
	return func(pool *workerPool[T]) {
		if maxWorker < 1 {
			return
		}
		pool.maxWorker = maxWorker
	}
}

func (wp *workerPool[T]) Handle(onSuccess func(resp <-chan T), onFailed func(errs <-chan error)) {
	wp.handler(func(resp <-chan T) {
		onSuccess(resp)
	}, func(errs <-chan error) {
		onFailed(errs)
	})
}

func (wp *workerPool[T]) Await() {
	wp.onClose()
}

func (wp *workerPool[T]) Submit(task Task[T]) {
	wp.wg.Add(1)
	go func() {
		wp.addTask(task)
	}()
}

func DoWithWorkerPool[T any](tasks []Task[T], opts ...WorkerPoolOption[T]) WorkerPool[T] {
	wp := defaultWorkerPool[T]()
	for _, opt := range opts {
		opt(wp)
	}
	wp.queuedTaskC = make(chan Task[T], wp.maxWorker)

	var (
		done = make(chan T, 1)
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

	wp.handler = func(onSuccess func(resp <-chan T), onFailed func(errs <-chan error)) {
		go func() {
			onSuccess(done)
		}()
		go func() {
			onFailed(fail)
		}()
	}

	return wp
}
