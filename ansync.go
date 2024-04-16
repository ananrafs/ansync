package ansync

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

type (
	Task   func() (interface{}, error)
	Action func() error
)
