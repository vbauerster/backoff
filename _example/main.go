package main

import (
	"fmt"
	"time"

	"github.com/vbauerster/backoff"
)

func main() {
	for i := 0; i < 5; i++ {
		d := backoff.DefaultStrategy.Backoff(i)
		fmt.Printf("%d: %v\n", i, d)
		time.Sleep(d)
	}

	fmt.Println()

	b := backoff.New(
		backoff.WithMaxDelay(300*time.Second),
		backoff.WithResetDelay(10*time.Second),
	)

	for i := 0; i < 10; i++ {
		if i > 0 && i%3 == 0 {
			time.Sleep(11 * time.Second)
		}
		d := b.Backoff(i + 1)
		fmt.Printf("%d: %v\n", i, d)
		time.Sleep(d)
	}
}
