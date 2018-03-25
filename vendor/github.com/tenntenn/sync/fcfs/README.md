# fcfs: First Come First Serve

## Document

see: https://godoc.org/github.com/tenntenn/sync/fcfs

## How to use

```
var g fcfs.Group
g.Go(func() (interface{}, error) {
    time.Sleep(10 * time.Second)
    return 100, nil
})

g.Go(func() (interface{}, error) {
    time.Sleep(1 * time.Second)
    return 200, nil
})

v, err := g.Wait()
if err != nil {
    log.Fatal(err)
}

fmt.Println(v)
```

## Contributors

* [tenntenn](https://github.com/tenntenn)
* [morikuni](https://github.com/morikuni)(base idea of algorithm)
