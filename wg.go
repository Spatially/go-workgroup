package workgroup

import (
	"fmt"
	"log"
	"runtime"
	"strings"
	"sync"
	"time"
)

var numCPU = runtime.NumCPU()

// Anything that can Perform() can be a Work.
type Work interface{}

//
type Worker func(int, Work)

// Called once per Worker
type Initializer func(int)

// This is the Work generator.
type Generator func(chan<- Work)

// A Generator function is a Source.
func (g Generator) Pump(wc chan<- Work) {
	g(wc)
}

// Sources of work implement Pump to send work on its provided channel.
type Source interface {
	Pump(chan<- Work)
}

//
type Monitor interface {
	Work() Work
	Elapsed() time.Duration
}

// The things a WorkGroup can do.
type WorkGroup interface {
	// A WorkGroup is a Source
	Source
	// A WorkGroup can Drain a Source
	Drain(Source) WorkGroup
	// A WorkGroup assigns a worker
	Configure(Initializer) WorkGroup
	// A WorkGroup assigns a worker
	With(Worker) WorkGroup
	// A WorkGroup can be started
	Go()
	//
	enablePump() WorkGroup
}

// The WorkGroup instance
type workGroup struct {
	sync.WaitGroup
	initialize struct {
		worker       sync.Once
		drain        sync.Once
		pump         sync.Once
		configurator sync.Once
	}
	in        chan Work
	out       chan Work
	options   Options
	worker    Worker
	configure Initializer
	name      string
	drain     func()
	monitor   chan time.Duration
}

// A *workGroup is a Source.
func (wg *workGroup) String() string {
	return wg.name
}

// A WorkGroup can Pump its Work ... if provided to another WorkGroup's Drain.
func (wg *workGroup) Pump(out chan<- Work) {
	for work := range wg.out {
		out <- work
	}
}

// Start the WorkGroup. For nested WorkGroups, this should be called only
// on the outer-most WorkGroup
func (wg *workGroup) Go() {

	worker := func(id int) {
		wg.Add(1)
		defer func() {
			wg.Done()
		}()
		if wg.configure != nil {
			wg.configure(id)
		}
		for {
			select {
			case assignment, ok := <-wg.in:
				if !ok {
					return
				}
				if wg.worker == nil {
					log.Fatal("ERROR: A Worker is mandatory: %+v : %+v", assignment, wg)
					return
				}

				// log.Printf("ASSIGNED: %+v : %+v", assignment, wg)
				started := time.Now()
				wg.worker(id, assignment)
				if wg.options.Timing {
					wg.monitor <- time.Since(started)
				}
				if wg.out != nil {
					wg.out <- assignment
				}
			case <-time.After(wg.options.Timeout):
				// TODO Provide a means of signalling the caller that this timed out.
				//      Probably using a function like With(), maybe OnTimeout().
				log.Printf("TIMEOUT: Waited %+v without seeing any work. %+v", wg.options.Timeout, wg)
			}
		}
	}

	for w := 0; w < wg.options.Workers; w++ {
		go worker(w)
	}

	wg.drain()
}

// Creates a WorkGroup with more than 1 running worker.
func FanOut(workers int, name ...string) WorkGroup {
	if workers == 1 {
		log.Println("It is not really a fan-out with only one worker.")
	}
	parameter := Options{Workers: workers}
	if name != nil {
		parameter.Name = strings.Join(name, "")
	}
	return New(parameter)
}

// Creates a WorkGroup with 1 running worker.
func FanIn(parameters ...Options) WorkGroup {
	var parameter Options
	if parameters == nil {
		parameter = Options{Workers: 1}
	} else if len(parameters) > 0 {
		parameter = parameters[0]
	}

	parameter.Workers = 1
	return New(parameter)
}

// For specifying parameters to New.
type Options struct {
	Workers int
	Name    string
	Timeout time.Duration
	Timing  bool
}

// If workers is <= 0, the number of workers will be set to the number of numCPU.
// When workers > 1, the workgroup is a fan-out. Maybe just use FanOut().
// When workers == 1, the workgroup is a fan-in. Maybe just use FanIn().
func New(parameters ...Options) WorkGroup {
	wg := &workGroup{
		WaitGroup: sync.WaitGroup{},
		in:        make(chan Work),
	}

	if parameters == nil {
		wg.options = Options{
			Workers: numCPU,
			Timeout: 0}
	} else if len(parameters) > 0 {
		wg.options = parameters[0]
	}

	if wg.options.Workers <= 0 || wg.options.Workers > 100*numCPU {
		wg.options.Workers = numCPU
	}
	if wg.options.Timeout <= 0 {
		wg.options.Timeout = 30 * time.Second
	}

	wg.name = fmt.Sprintf("%s/%p/%d", wg.options.Name, &wg.WaitGroup, wg.options.Workers)
	if wg.options.Timing {
		wg.monitor = make(chan time.Duration, 0)
	}

	return wg
}

// Initializes a WorkGroup's Worker.
func (wg *workGroup) Configure(initializer Initializer) WorkGroup {
	if initializer == nil {
		return nil
	}
	wg.initialize.configurator.Do(func() { wg.configure = initializer })
	return wg
}

// Initializes a WorkGroup's Worker.
func (wg *workGroup) With(worker Worker) WorkGroup {
	if worker == nil {
		log.Fatal("ERROR: A Worker is mandatory: %+v", wg)
		return nil
	}
	wg.initialize.worker.Do(func() { wg.worker = worker })
	return wg
}

// Private function to enable a WorkGroup's Pump.
func (wg *workGroup) enablePump() WorkGroup {
	wg.initialize.pump.Do(func() { wg.out = make(chan Work) })
	return wg
}

// Initializes the WorkGroup's Drain.
func (wg *workGroup) Drain(source Source) WorkGroup {
	if source == nil {
		log.Fatal("ERROR: A Source is mandatory: %+v", wg)
		return nil
	}

	wg.initialize.drain.Do(func() {

		// If the source is a WorkGroup, then configure the source-WorkGroup so
		// that it will Pump its Work.
		var sourceWorkGroup WorkGroup
		switch sg := source.(type) {
		case WorkGroup:
			sourceWorkGroup = sg.enablePump()
		}

		// Assign the drain function that'll be called by Go.
		wg.drain = func() {

			// If the soruce is a WorkGroup, get it going.
			if sourceWorkGroup != nil {
				go sourceWorkGroup.Go()
			}

			data := make(chan Work)
			go func() {
				defer close(data)
				source.Pump(data)
			}()

			for work := range data {
				wg.in <- work
			}
			close(wg.in)
			wg.Wait()
			// TODO is this placement correct or should it be above wg.Wait()?
			if wg.out != nil {
				close(wg.out)
			}
		}
	})

	return wg
}
