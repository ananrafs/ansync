# Ansync

Anysnc is tools related to goroutine, perform cancellable operation/task, workerpool implementation, etc




## cancellation
### example
```go
import (
	"context"
	"fmt"
	"github.com/ananrafs/ansync/cancellation"
	"time"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	err := cancellation.Do(ctx, func() error {
		for i := 0; i < 10; i++ {
			fmt.Printf("%d\n", i)
			time.Sleep(time.Second)
		}
		return nil
	})
	if err != nil {
		panic(err)
	}
}
```
will panic after 2 seconds

## workerpool
### example
```go
import (
        "context"
        "fmt"
        "github.com/ananrafs/ansync/workerpool"
        "github.com/ananrafs/ansync"
        "time"
)

func main() {
	handle, closer := workerpool.Do(
		[]ansync.Task{
			// define task to perform
			func() (interface{}, error) {
				for i := 0; i < 10; i++ {
					fmt.Printf("%d from task 1 \n", i)
					time.Sleep(time.Second)
				}
				return 12, nil
			},
			func() (interface{}, error) {
				for i := 0; i < 10; i++ {
					fmt.Printf("%d from task 2 \n", i*2)
					time.Sleep(2 * time.Second)
				}
				return nil, errors.New("matatata matuta")
			},
			func() (interface{}, error) {
				for i := 0; i < 10; i++ {
					fmt.Printf("%d from task 3 \n", i*3)
					time.Sleep(3 * time.Second)
				}
				return 32, nil
			},
		},
		workerpool.WithMaxWorker(2),
	)
	// need to defer closer
	defer closer()

	// handling response and error
	handle(func(res <-chan interface{}) {
		for r := range res {
			fmt.Println(r)
		}
	}, func(errors <-chan error) {
		for err := range errors {
			fmt.Println(err)
		}
	})
}
```