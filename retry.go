package ansync

import (
	"fmt"
	"time"
)

type (
	retrySetup struct {
		maxRetries int
		delayer    func(int) time.Duration

		hook retryHook
	}

	retryHook struct {
		onFailure func(error) error
		onSuccess func() error
	}

	RetryOption func(r *retrySetup)
)

var (
	maxAttemptErr = fmt.Errorf("max attempts exceeded")
)

func newDefaultRetrySetup(maxRetries int, delayer func(int) time.Duration) *retrySetup {
	return &retrySetup{
		maxRetries: maxRetries,
		delayer:    delayer,
	}
}

func WithRetry(maxRetries int) RetryOption {
	return func(r *retrySetup) {
		r.maxRetries = maxRetries
	}
}

func WithDelay(delayer func(attempt int) time.Duration) RetryOption {
	return func(r *retrySetup) {
		r.delayer = delayer
	}
}

func WithOnFailure(onFailure func(error) error) RetryOption {
	return func(r *retrySetup) {
		r.hook.onFailure = onFailure
	}
}

func WithOnSuccess(onSuccess func() error) RetryOption {
	return func(r *retrySetup) {
		r.hook.onSuccess = onSuccess
	}
}

func DoActionWithRetry(act Action, opts ...RetryOption) error {
	setup := newDefaultRetrySetup(1, func(retryAttempt int) time.Duration {
		return 1 * time.Second
	})

	for _, opt := range opts {
		opt(setup)
	}

	var (
		retryAttempt = 0
		err          error
	)

	for {
		err = act()
		if err == nil {
			if setup.hook.onSuccess != nil {
				return setup.hook.onSuccess()
			}
			return nil
		}
		if retryAttempt >= setup.maxRetries {
			if setup.hook.onFailure != nil {
				return setup.hook.onFailure(err)
			}
			return maxAttemptErr
		}
		if setup.delayer != nil {
			time.Sleep(setup.delayer(retryAttempt))
		}
		retryAttempt++
	}
}

func DoTaskWithRetry[T any](task Task[T], opts ...RetryOption) (result T, err error) {
	setup := newDefaultRetrySetup(1, func(retryAttempt int) time.Duration {
		return 1 * time.Second
	})

	for _, opt := range opts {
		opt(setup)
	}

	var (
		done = make(chan T)
		fail = make(chan error)
	)

	go func(setup *retrySetup) {
		var (
			retryAttempt = 0
			err          error
			resp         T
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
		return result, err
	}

}
