package main

import (
	"log"
	"math/rand"
	"runtime"
	"sync"
	"time"

	workgroup "github.com/Spatially/go-workgroup"
)

// A comparison of workgroup vs pure sync.WaitGroup
func main() {

	workers := runtime.NumCPU()
	runtime.GOMAXPROCS(workers)

	generate := 100

	group := sync.WaitGroup{}
	group.Add(2)
	go func() {
		usingWorkgroup(workers, generate)
		group.Done()
	}()
	go func() {
		usingWaitGroup(workers, generate)
		group.Done()
	}()
	group.Wait()
}

// The workgroup version
func usingWorkgroup(workers, generate int) {

	// This is a Worker function. The workgroup will start however many of these
	// you specify. In this example, it will start one for each CPU (see below).
	workhorse := func(worker int, work workgroup.Work) {
		time.Sleep(time.Duration(rand.Int63n(1000)) * time.Millisecond)
		log.Printf("workgroup: worker %d handing work %d.", worker, work)
	}

	// This is a Work-Generator. It simply feeds work to each Worker goroutine
	// as each is ready for Work. Although the Workers are goroutines, a
	// workgroup uses sync.WaitGroup interanlly so this goroutine will block
	// on the out channel until a Worker reads from the channel.
	// The completion of this signals the workgroup's cleanup process (all the
	// Workers will complete their work.)
	workUnits := workgroup.Generator(func(out chan<- workgroup.Work) {
		for i := 0; i < generate; i++ {
			out <- i
		}
	})

	// This configures and initates the workgroup.
	// FanOut specifies how many Workers to start.
	// Drain specifies the Generator function.
	// With provides the Worker function.
	workgroup.FanOut(workers).Drain(workUnits).With(workhorse).Go()

	// This is just to show that Go blocks.
	log.Printf("Done: elapsed:+%v", time.Now())

}

// Another equivalent workgroup version but optimized for space, shaving two lines from above's.
func usingOptimizedWorkgroup(workers, generate int) {

	workgroup.FanOut(workers).Drain(workgroup.Generator(func(out chan<- workgroup.Work) {
		for i := 0; i < generate; i++ {
			out <- i
		}
	})).With(func(worker int, work workgroup.Work) {
		time.Sleep(time.Duration(rand.Int63n(1000)) * time.Millisecond)
		log.Printf("workgroup: worker %d handing work %d.", worker, work)
	}).Go()

	// This is just to show that Go blocks.
	log.Printf("Done: elapsed:+%v", time.Now())

}

// The equivalent sync.WaitGroup version.
func usingWaitGroup(workers, generate int) {

	var wg sync.WaitGroup

	in := make(chan int)

	wg.Add(workers)
	for w := 0; w < workers; w++ {
		go func(worker int, in <-chan int) {
			defer wg.Done()

			// This is essentially the workgroup Worker function.
			for work := range in {
				time.Sleep(time.Duration(rand.Int63n(1000)) * time.Millisecond)
				log.Printf("WaitGroup: worker %d handing work %d.", worker, work)
			}

		}(w, in)
	}

	// This is essentially the workgroup Generator function.
	for w := 0; w < generate; w++ {
		in <- w
	}

	close(in)

	// This is essentially the workgroup Go() call.
	wg.Wait()

	log.Printf("Done: elapsed:+%v", time.Now())
}
