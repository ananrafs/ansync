package ansync

import (
	"context"
)

func DoActionWithCancellation(ctx context.Context, act Action) error {
	done := make(chan error, 1)
	go func() {
		err := act()
		done <- err
		close(done)
	}()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case ret := <-done:
		return ret
	}
}

func DoTaskWithCancellation[T any](ctx context.Context, task Task[T]) (result T, err error) {
	done := make(chan T, 1)
	fail := make(chan error, 1)

	go func() {
		res, err := task()
		if err != nil {
			fail <- err
			close(fail)
			close(done)
			return
		}
		done <- res
		close(done)
		close(fail)
	}()

	select {
	case <-ctx.Done():
		return result, ctx.Err()
	case ret := <-done:
		return ret, nil
	case ret := <-fail:
		return result, ret
	}
}
