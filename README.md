workgroup - wraps sync.WaitGroup
=======================================

Package `workgroup` is a [Go](http://golang.org) client library providing a wrapper for implementing a sync.WaitGroup. It attempts to eliminate all the boilerplate often required when writing sync.WaitGroup workers.

# Status

[![Build Status](https://travis-ci.org/Spatially/go-workgroup.png?branch=master)](https://travis-ci.org/Spatially/go-workgroup)


# Installation

```
go get -v github.com/Spatially/go-workgroup
```


# Documentation

See [GoDoc](http://godoc.org/github.com/Spatially/go-workgroup) or [Go Walker](http://gowalker.org/github.com/Spatially/go-workgroup) for automatically generated documentation.


# Usage

### 1. Define a Worker

This is a Worker function. The workgroup will start however many of these you specify. In this example, it will start one for each CPU (see 3 below).

```go
workhorse := func(worker int, work workgroup.Work) {
	log.Printf("%d %+v Done : +%v", worker, work, time.Now())
}
```

### 2. Define a Work-Generator function

This is a Work-Generator. It simply feeds work to each `workhorse` goroutine as each is ready for Work. Although the Workers are goroutines, a workgroup uses sync.WaitGroup internally so this goroutine will block the out channel until a Worker reads from the channel. The completion of this signals the workgroup's cleanup process (all the Workers will complete their work).

```go
workUnits := workgroup.Generator(func(out chan<- workgroup.Work) {
	for i := 0; i < 100; i++ {
		out <- i
	}
})
```

### 3. Initiate the WorkGroup with 8 workers

This configures and initates the workgroup.
- `FanOut` specifies how many Workers to start. 8 in this case.
- `Drain` specifies the Generator function.
- `With` provides the Worker function.

```go
workgroup.FanOut(8).Drain(workUnits).With(workhorse).Go()
```
