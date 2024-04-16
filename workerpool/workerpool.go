package workerpool

import (
	"sync"
)

type (
	workerPool struct {
		maxWorker   int
		action      func(interface{}) (interface{}, error)
		queuedTaskC chan WorkerAction
		wg          sync.WaitGroup
	}

	Options      func(*workerPool)
	WorkerAction func() (interface{}, error)
)

func (wp *workerPool) addTask(task WorkerAction) {
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

func Do(actions []WorkerAction, opts ...Options) (responseHandler func(func(<-chan interface{}), func(<-chan error)), closer func()) {
	wp := defaultWorkerPool()
	for _, opt := range opts {
		opt(wp)
	}
	wp.queuedTaskC = make(chan WorkerAction, wp.maxWorker)

	var (
		done = make(chan interface{}, 1)
		fail = make(chan error, 1)
	)

	wp.run(done, fail)
	wp.wg.Add(len(actions))

	go func() {
		for _, task := range actions {
			wp.addTask(task)
		}
	}()

	return func(onSuccess func(<-chan interface{}), onFailed func(<-chan error)) {
			go func() {
				onSuccess(done)
			}()
			go func() {
				onFailed(fail)
			}()
		}, func() {
			wp.wg.Wait()
			wp.close(func() {
				close(done)
				close(fail)
			})
		}

}

//// EXAMPLE
//
//func main() {
//	handle, closer := workerpool.Do(
//		[]workerpool.WorkerAction{
//			// define task to perform
//			func() (interface{}, error) {
//				for i := 0; i < 10; i++ {
//					fmt.Printf("%d from task 1 \n", i)
//					time.Sleep(time.Second)
//				}
//				return 12, nil
//			},
//			func() (interface{}, error) {
//				for i := 0; i < 10; i++ {
//					fmt.Printf("%d from task 2 \n", i*2)
//					time.Sleep(2 * time.Second)
//				}
//				return nil, errors.New("matatata matuta")
//			},
//			func() (interface{}, error) {
//				for i := 0; i < 10; i++ {
//					fmt.Printf("%d from task 3 \n", i*3)
//					time.Sleep(3 * time.Second)
//				}
//				return 32, nil
//			},
//		},
//		workerpool.WithMaxWorker(2),
//	)
// 	// need to defer closer
//	defer closer()

// 	// handling response and error
//	handle(func(res <-chan interface{}) {
//		for r := range res {
//			fmt.Println(r)
//		}
//	}, func(errors <-chan error) {
//		for err := range errors {
//			fmt.Println(err)
//		}
//	})
//}
