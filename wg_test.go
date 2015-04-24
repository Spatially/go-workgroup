package workgroup

import (
	"fmt"
	"log"
	"math/rand"
	"runtime"
	"testing"
	"time"
)

// This is Work
type T struct {
	*testing.T
	i      int
	result string
}

//
func helpProcess(t *testing.T, workers int) {

	runtime.GOMAXPROCS(runtime.NumCPU())

	start := time.Now()

	New(Options{Workers: workers}).Drain(Generator(func(out chan<- Work) {
		for i := 0; i < 100; i++ {
			out <- T{t, i, ""}
		}
	})).With(func(worker int, work Work) {
		switch w := work.(type) {
		case T:
			sleeper := time.Duration(rand.Int31n(900)+100) * time.Millisecond
			time.Sleep(sleeper)
			w.result = fmt.Sprintf("%d %d Done: %+v :+%v", worker, w.i, sleeper, time.Now())
			t.Log(w.result)
		}
	}).Go()

	t.Logf("Done: elapsed:+%v", time.Now().Sub(start))
}

//
func TestProcessCPUs(t *testing.T) {
	helpProcess(t, 0)
}

//
func TestProcess50(t *testing.T) {
	helpProcess(t, 50)
}

//
func TestProcess500(t *testing.T) {
	helpProcess(t, 500)
}

//
func TestTimeout(t *testing.T) {

	runtime.GOMAXPROCS(runtime.NumCPU())

	start := time.Now()

	sleeper := func(n int32) time.Duration {
		d := time.Duration(rand.Int31n(n)) * time.Millisecond
		time.Sleep(d)
		return d
	}

	fanOutWorker := func(worker int, work Work) {
		// Using log.Printf instead of t.Logf so that the timeouts are
		// interlaced with these logs to better see where the timeouts
		// occur. NOTE: Once a timeout-handler is implmented, this'll
		// no longer be "necessary".
		log.Printf("%d %+v Done: %+v :+%v", worker, work, sleeper(500), time.Now())
	}
	workUnits := Generator(func(out chan<- Work) {
		for i := 0; i < 100; i++ {
			out <- i
			if i%13 == 0 {
				sleeper(500)
			}
		}
	})

	New(Options{Timeout: 100 * time.Millisecond}).Drain(workUnits).With(fanOutWorker).Go()

	t.Logf("Done: elapsed:+%v", time.Now().Sub(start))
}

//
func TestElapsed(t *testing.T) {

	runtime.GOMAXPROCS(runtime.NumCPU())

	start := time.Now()

	sleeper := func(n int32) time.Duration {
		d := time.Duration(rand.Int31n(n)) * time.Millisecond
		time.Sleep(d)
		return d
	}

	fanOutWorker := func(worker int, work Work) {
		log.Printf("%d %+v Done: %+v :+%v", worker, work, sleeper(500), time.Now())
	}
	workUnits := Generator(func(out chan<- Work) {
		for i := 0; i < 100; i++ {
			out <- i
			if i%13 == 0 {
				sleeper(500)
			}
		}
	})

	New(Options{Timeout: 100 * time.Millisecond}).Drain(workUnits).With(fanOutWorker).Go()

	t.Logf("Done: elapsed:+%v", time.Now().Sub(start))
}

//
func makeFanInWorker() Worker {
	sum := 0
	return func(worker int, work Work) {
		switch t := work.(type) {
		case T:
			sum += t.i
			t.Logf("%d %+v %+d Fan-in Done: %+v", worker, work, sum, time.Now())
		}
	}
}

//
func TestChain(t *testing.T) {
	runtime.GOMAXPROCS(runtime.NumCPU())

	start := time.Now()

	fanOutWorker := func(worker int, work Work) { t.Logf("%d %+v Fan-Out Done: %+v", worker, work, time.Now()) }
	workUnits := Generator(func(out chan<- Work) {
		for i := 0; i < 100; i++ {
			out <- T{t, i, ""}
		}
	})

	fanOut := FanOut(8).Drain(workUnits).With(fanOutWorker)
	FanIn().Drain(fanOut).With(makeFanInWorker()).Go()

	t.Logf("Done: elapsed:+%v", time.Now().Sub(start))
}

type cT struct {
	i int
	t time.Time
}

// This is Work
type Tc struct {
	*testing.T
	i      int
	result string
}

//
func TestConfiguration(t *testing.T) {
	runtime.GOMAXPROCS(runtime.NumCPU())

	start := time.Now()

	workers := 8
	configs := make([]cT, workers)

	fanOutWorker := func(worker int, work Work) { t.Logf("%d %+v Fan-Out Done: %+v", worker, work, time.Now()) }
	workUnits := Generator(func(out chan<- Work) {
		for i := 0; i < 16; i++ {
			out <- Tc{t, i, ""}
		}
	})

	configurator := func(id int) {
		configs[id] = cT{id, time.Now()}
	}

	FanOut(workers).Configure(configurator).Drain(workUnits).With(fanOutWorker).Go()

	t.Logf("Done: elapsed:+%v", time.Now().Sub(start))
}

/*













*/
