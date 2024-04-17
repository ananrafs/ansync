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
    "errors"
    "fmt"
    "github.com/ananrafs/ansync"
    "github.com/ananrafs/ansync/workerpool"
    "time"
)

func main() {
    wp := workerpool.Do(
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
                        fmt.Printf("%d from task 2 \n", i)
                        time.Sleep(time.Second)
                    }
                    return nil, errors.New("matatata matuta")
                },
                func() (interface{}, error) {
                    for i := 0; i < 10; i++ {
                        fmt.Printf("%d from task 3 \n", i)
                        time.Sleep(time.Second)
                    }
                    return "natsu", nil
                },
            },
                workerpool.WithMaxWorker(2),
        )
        // need to await
        defer wp.Await()
        
	// submit new task
        wp.Submit(func() (interface{}, error) {
            for i := 0; i < 10; i++ {
                fmt.Printf("%d from task 4 \n", i)
                time.Sleep(time.Second)
            }
            return "oh my god", nil
        })
        
        wp.Submit(func() (interface{}, error) {
            for i := 0; i < 10; i++ {
                fmt.Printf("%d from task 5 \n", i)
                time.Sleep(time.Second)
            }
            return "ululu", nil
        })
        
        // handling response and error
        wp.Handle(func(res <-chan interface{}) {
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