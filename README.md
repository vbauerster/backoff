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
```

### Output
```
0: 1s
1: 1.533129183s
2: 2.428043674s
3: 4.696643849s
4: 6.439863432s

0: 1.864278011s
1: 2.139655742s
2: 4.461737515s
3: 1.55156559s
4: 2.288627626s
5: 3.649828801s
6: 1.665687677s
7: 2.447051421s
8: 3.278575388s
9: 1.531698143s
```
