package ansync

type (
	Pipe[T any]          func(in <-chan T) (out chan T, err error)
	pipelineSetup[T any] struct {
	}
	PipelineOption[T any] func(*pipelineSetup[T])
)

func defaultPipelineSetup[T any]() *pipelineSetup[T] {
	return &pipelineSetup[T]{}
}

func DoWithPipeline[T any](stream chan T, pipes []Pipe[T], opts ...PipelineOption[T]) (out chan T, err error) {
	setup := defaultPipelineSetup[T]()
	for _, opt := range opts {
		opt(setup)
	}
	streamCh := stream
	for _, pipe := range pipes {
		currentStream := streamCh
		streamCh, err = pipe(currentStream)
		if err != nil {
			return nil, err
		}

	}

	return streamCh, nil
}
