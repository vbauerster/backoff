# Exponential Backoff

The algorithm extracted from [grpc](https://github.com/grpc/grpc/blob/master/doc/connection-backoff.md) library
and augmented with functional options.

#### [Example](_example/main.go)
```go
	for i := 0; i < 5; i++ {
		d := backoff.DefaultStrategy.Backoff(i)
		fmt.Printf("%d: %v\n", i, d)
		time.Sleep(d)
	}

	b := backoff.New(
		backoff.WithBaseDelay(2*time.Second),
		backoff.WithMaxDelay(300*time.Second),
		backoff.WithResetDelay(10*time.Second),
	)

	for i := 0; i < 10; i++ {
		if i > 0 && i%3 == 0 {
			time.Sleep(11 * time.Second)
		}
		d := b.Backoff(i)
		fmt.Printf("%d: %v\n", i, d)
		time.Sleep(d)
	}
```

### Output
```
0: 1s
1: 1.861015984s
2: 2.894111824s
3: 3.529445921s
4: 5.502962438s
0: 2s
1: 2.716279239s
2: 5.724163401s
3: 2s
4: 3.469467053s
5: 4.197908886s
6: 2s
7: 3.479944348s
8: 5.800812833s
9: 2s

```
