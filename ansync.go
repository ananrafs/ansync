package ansync

type (
	Task[T any] func() (T, error)
	Action      func() error
)
