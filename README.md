# Exponential Backoff

The algorithm extracted from [grpc](https://github.com/grpc/grpc/blob/master/doc/connection-backoff.md) library
and augmented with functional options. 

#### [Example](_example/main.go)
```go
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
```

### Output
```
1s
1.390532561s
2.520493227s
4.805583273s
5.899954303s
9.598781269s
18.465925964s
31.532035436s
36.411486205s
55.91100951s
2s
3.591917239s
5.84878073s
9.201170154s
11.855413451s
21.331910845s
36.474201007s
59.198877125s
1m18.302021863s
2m43.010456169s
```
