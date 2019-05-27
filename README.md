# go-vector [![GoDoc](https://godoc.org/github.com/erizocosmico/go-vector?status.svg)](https://godoc.org/github.com/erizocosmico/go-vector) [![Build Status](https://travis-ci.org/erizocosmico/go-vector.svg?branch=master)](https://travis-ci.org/erizocosmico/go-vector) [![codecov](https://codecov.io/gh/erizocosmico/go-vector/branch/master/graph/badge.svg)](https://codecov.io/gh/erizocosmico/go-vector) [![Go Report Card](https://goreportcard.com/badge/github.com/erizocosmico/go-vector)](https://goreportcard.com/report/github.com/erizocosmico/go-vector) [![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

Persistent bit-partitioned vector trie implementation for Go, a slice-like persistent data structure. Pretty much like the one used by the Clojure language.

## Install

```
go get github.com/erizocosmico/go-vector
```

## Usage

```go
vemtpy := vector.New() // empty vector
v := vector.New(1, 2, 3, 4, 5) // vector with items

v = v.Append(6) // new vector with 6 appended at the end

elem := v.Get(2) // elem is 3

v.Last() // last element
v.First() // first element
v.Tail() // new vector without the first element

v = v.Set(0, -1) // Set element 0 to -1

// Iterate over all elements.
err := v.Range(func(x interface{}) error {
    // do something with x
    return nil
})

v.Slice() // return the elements as a slice

squared := v.Map(func(x interface{}) interface{} {
    x := x.(int)
    return x * x
})

even := v.Filter(func(x interface{}) bool {
    return x.(int) % 2 == 0
})

firstThree := v.Take(3)
allButFirst := v.Drop(1)

vector.Equal(New(-1, 2, 3, 4, 5), v) // will output true

// Check if they're equal using a custom function. This will return false
// because the vectors are not equal.
vector.EqualFunc(New(1, 2, 3, 4, 5), v, func(a, b interface{}) bool {
    return a.(int) == b.(int)
})
```

For more info, check out [the package documentation](https://godoc.org/github.com/erizocosmico/go-vector).

## Benchmarks

```
BenchmarkAppend/10-4         	 3000000	       547 ns/op	     423 B/op	       4 allocs/op
BenchmarkAppend/100-4        	 3000000	       517 ns/op	     423 B/op	       4 allocs/op
BenchmarkAppend/1000-4       	 3000000	       502 ns/op	     423 B/op	       4 allocs/op
BenchmarkGet/10-4            	100000000	        20.8 ns/op	       0 B/op	       0 allocs/op
BenchmarkGet/100-4           	100000000	        20.8 ns/op	       0 B/op	       0 allocs/op
BenchmarkGet/1000-4          	100000000	        20.8 ns/op	       0 B/op	       0 allocs/op
BenchmarkSet/10-4            	10000000	       236 ns/op	     248 B/op	       4 allocs/op
BenchmarkSet/100-4           	 3000000	       567 ns/op	    1104 B/op	       5 allocs/op
BenchmarkSet/1000-4          	 2000000	       726 ns/op	    1136 B/op	       5 allocs/op
```

## License

MIT, see [LICENSE](/LICENSE)