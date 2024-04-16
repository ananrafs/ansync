package cancellation

import (
	"context"
	"github.com/ananrafs/ansync"
)

func Do(ctx context.Context, act ansync.Action) error {
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

func DoWithReturn(ctx context.Context, task ansync.Task) (interface{}, error) {
	done := make(chan interface{}, 1)
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
		return nil, ctx.Err()
	case ret := <-done:
		return ret, nil
	case ret := <-fail:
		return nil, ret
	}
}
