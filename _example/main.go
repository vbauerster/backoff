package main

import (
	"fmt"
	"time"

	"github.com/vbauerster/backoff"
)

func main() {
	b := backoff.DefaultStrategy
	for i := 0; i < 10; i++ {
		fmt.Println(b.Backoff(i))
	}

	b = backoff.New(
		backoff.WithBaseDelay(2*time.Second),
		backoff.WithMaxDelay(300*time.Second),
	)
	for i := 0; i < 10; i++ {
		fmt.Println(b.Backoff(i))
	}
}
