package ansync

import (
	"time"
)

type (
	retrySetup struct {
		maxRetries int
		delayer    func(int) time.Duration
	}

	RetryOptions func(r *retrySetup)
)

func newDefaultRetrySetup(maxRetries int, delayer func(int) time.Duration) *retrySetup {
	return &retrySetup{
		maxRetries: maxRetries,
		delayer:    delayer,
	}
}

func WithRetry(maxRetries int) RetryOptions {
	return func(r *retrySetup) {
		r.maxRetries = maxRetries
	}
}

func WithDelay(delayer func(int) time.Duration) RetryOptions {
	return func(r *retrySetup) {
		r.delayer = delayer
	}
}

func DoActionWithRetry(act Action, opts ...RetryOptions) error {
	setup := newDefaultRetrySetup(1, func(retryAttempt int) time.Duration {
		return 1 * time.Second
	})

	for _, opt := range opts {
		opt(setup)
	}

	var (
		done = make(chan interface{})
	)

	go func(setup *retrySetup) {
		var (
			retryAttempt = 0
			err          error
		)
		defer func() {
			if err != nil {
				done <- err
			} else {
				done <- struct{}{}
			}
		}()

		for {
			err = act()
			if err == nil {
				return
			}
			if retryAttempt >= setup.maxRetries {
				return
			}
			if setup.delayer != nil {
				time.Sleep(setup.delayer(retryAttempt))
			}
			retryAttempt++
		}
	}(setup)

	res := <-done
	if err, ok := res.(error); ok {
		return err
	}

	return nil
}

func DoTaskWithRetry(task Task, opts ...RetryOptions) (interface{}, error) {
	setup := newDefaultRetrySetup(1, func(retryAttempt int) time.Duration {
		return 1 * time.Second
	})

	for _, opt := range opts {
		opt(setup)
	}

	var (
		done = make(chan interface{})
		fail = make(chan error)
	)

	go func(setup *retrySetup) {
		var (
			retryAttempt = 0
			err          error
			resp         interface{}
		)
		defer func() {
			if err != nil {
				fail <- err
			} else {
				done <- resp
			}
		}()

		for {
			resp, err = task()
			if err == nil {
				return
			}
			if retryAttempt >= setup.maxRetries {
				return
			}
			if setup.delayer != nil {
				time.Sleep(setup.delayer(retryAttempt))
			}
			retryAttempt++
		}
	}(setup)

	select {
	case response := <-done:
		return response, nil
	case err := <-fail:
		return nil, err
	}

}
