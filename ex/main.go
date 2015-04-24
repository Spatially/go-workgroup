package main

import (
	"log"
	"runtime"
	"time"

	workgroup "github.com/Urban4M/go-workgroup"
)

// A workgroup example.
func main() {

	// This is a Worker function. The workgroup will start however many of these
	// you specify. In this example, it will start one for each CPU (see below).
	fanOutWorker := func(worker int, work workgroup.Work) {
		log.Printf("%d %+v Done : +%v", worker, work, time.Now())
	}

	// This is a Work-Generator. It simply feeds work to each Worker goroutine
	// as each is ready for Work. Although the Workers are goroutines, a
	// workgroup uses sync.WaitGroup interanlly so this goroutine will block
	// on the out channel until a Worker reads from the channel.
	// The completion of this signals the workgroup's cleanup process (all the
	// Workers will complete their work.)
	workUnits := workgroup.Generator(func(out chan<- workgroup.Work) {
		for i := 0; i < 100; i++ {
			out <- i
		}
	})

	workers := runtime.NumCPU()
	runtime.GOMAXPROCS(workers)

	// This configures and initates the workgroup.
	// FanOut specifies how many Workers to start.
	// Drain specifies the Generator function.
	// With provides the Worker function.
	workgroup.FanOut(workers).Drain(workUnits).With(fanOutWorker).Go()

	// This is just to show that Go blocks.
	log.Printf("Done: elapsed:+%v", time.Now())
}
