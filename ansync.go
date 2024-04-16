package ansync

type (
	Task   func() (interface{}, error)
	Action func() error
)
