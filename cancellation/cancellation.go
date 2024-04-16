package cancellation

import "context"

func Do(ctx context.Context, f func() error) error {
	done := make(chan error, 1)
	go func() {
		err := f()
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

func DoWithReturn(ctx context.Context, f func() (interface{}, error)) (interface{}, error) {
	done := make(chan interface{}, 1)
	fail := make(chan error, 1)

	go func() {
		res, err := f()
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
