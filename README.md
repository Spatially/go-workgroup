workgroup - wraps sync.WaitGroup
=======================================

Package `workgroup` is a [Go](http://golang.org) client library providing a wrapper for implementing a WaitGroup.


# Status

[![Build Status](https://travis-ci.org/Urban4M/go-workgroup.png?branch=master)](https://travis-ci.org/Urban4M/go-workgroup)


# Installation

```
go get -v github.com/Urban4M/go-workgroup
```


# Documentation

See [GoDoc](http://godoc.org/github.com/Urban4M/go-workgroup) or [Go Walker](http://gowalker.org/github.com/Urban4M/go-workgroup) for automatically generated documentation.


# Usage

### 1. Define a Work-Generator function

```go
workUnits := workgroup.Generator(func(out chan<- workgroup.Work) {
	for i := 0; i < 100; i++ {
		out <- i
	}
})
```

### 2. Define a Worker

```go
worker := func(worker int, work workgroup.Work) {
	log.Printf("%d %+v Done : +%v", worker, work, time.Now())
}
```

### 3. Initiate the WorkGroup with 8 workers

```go
workgroup.FanOut(8).Drain(workUnits).With(fanOutWorker).Go()
```
