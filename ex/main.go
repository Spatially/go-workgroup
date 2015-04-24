package main

import (
	"log"
	"runtime"
	"time"

	workgroup "github.com/Urban4M/go-workgroup"
)

//
func main() {

	runtime.GOMAXPROCS(runtime.NumCPU())

	fanOutWorker := func(worker int, work workgroup.Work) {
		log.Printf("%d %+v Done : +%v", worker, work, time.Now())
	}

	workUnits := workgroup.Generator(func(out chan<- workgroup.Work) {
		for i := 0; i < 100; i++ {
			out <- i
		}
	})

	workgroup.FanOut(0).Drain(workUnits).With(fanOutWorker).Go()

	log.Printf("Done: elapsed:+%v", time.Now())
}
