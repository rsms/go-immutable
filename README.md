# Immutable data structures for Go

[![GitHub tag (latest SemVer)](https://img.shields.io/github/tag/rsms/go-immutable.svg)][godoc]
[![PkgGoDev](https://pkg.go.dev/badge/github.com/rsms/go-immutable)][godoc]
[![Go Report Card](https://goreportcard.com/badge/github.com/rsms/go-immutable)](https://goreportcard.com/report/github.com/rsms/go-immutable)

[godoc]: https://pkg.go.dev/github.com/rsms/go-immutable

- Based on a immutable persistent Hash Array Mapped Trie (HAMT)
- Minimal interface to the core HAMT implementation so that you can easily
  implement your own structures on top of it.
- Comes with some set and map implementations ready to go
- Inspired by Clojure's data structures
- Excellent performance (lookup is about the same as native go maps,
  insertion about 20% that of mutable native go maps.)

Example:

```go
package main

import (
  "fmt"
  "github.com/rsms/go-immutable"
)

m := EmptyStrMap

m1 := m.Set("Hello", 123)
m2 := m.Set("Hello", 456).Set("Sun", 9)
m3 := m2.Del("Hello")

fmt.Printf("m1: %s\n", m1)
fmt.Printf("m2: %s\n", m2)
fmt.Printf("m3: %s\n", m3)

// Output:
// m1: {"Hello": 123}
// m2: {"Sun": 9, "Hello": 456}
// m3: {"Sun": 9}
```

## Benchmark

```
TEST                            SAMPLES       TIME PER ITERATION
BenchmarkHamtLookup_10          48042776          23.8 ns/op
BenchmarkHamtLookup_100         36947116          31.0 ns/op
BenchmarkHamtLookup_1000        32566189          34.8 ns/op
BenchmarkHamtLookup_10000       24775086          49.1 ns/op

BenchmarkGoMapLookup_10         69836985          16.3 ns/op
BenchmarkGoMapLookup_100        66960720          16.5 ns/op
BenchmarkGoMapLookup_1000       36865258          30.7 ns/op
BenchmarkGoMapLookup_10000      28110822          43.3 ns/op

BenchmarkHamtInsert_10           8019391         137.0 ns/op
BenchmarkHamtInsert_25           6851881         174.0 ns/op
BenchmarkHamtInsert_50           5419015         221.0 ns/op
BenchmarkHamtInsert_100          3860596         293.0 ns/op
BenchmarkHamtInsert_1000         2596484         469.0 ns/op
BenchmarkHamtInsert_5000         1892996         644.0 ns/op

BenchmarkGoMapInsert_10         17940093          66.5 ns/op
BenchmarkGoMapInsert_25         13996856          74.3 ns/op
BenchmarkGoMapInsert_50         14645752          83.3 ns/op
BenchmarkGoMapInsert_100        14309284          82.8 ns/op
BenchmarkGoMapInsert_1000       10754890         106.0 ns/op
BenchmarkGoMapInsert_5000       11948517         100.0 ns/op
```

Results are from a 2018 MacBook Pro.
`BenchmarkHamt*` and `BenchmarkGo*` are tests with same input using HAMT and native Go maps,
respectively.
Run these benchmarks yourself with `./dev -bench`.

Keep in mind that HAMT is immutable and derivative data structure which requires lots of
memory allocations, compared to the mutable in-place native go maps.
I've chosen to compare the HAMT implementation with Go maps since Go maps are likely what
you are familiar with. :-)
