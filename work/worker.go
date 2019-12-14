package work


import (
	"fmt"
	"time"
	"sync"
)

// ProcessorFunc signature that defines the dependency injection to process "Jobs"
type ProcessorFunc func(resource interface {}) error

// ResultProcessorFunc signature that defines the dependency injection to process "Results"
type ResultProcessorFunc func(result Result) error


// Job Structure that wraps Jobs information
type Job struct {

	id int
	resource interface{}
}

// Result holds the main structure for worker processed job results.
type Result struct {
	Job Job
	Err error
}

// Manager generic struct that keeps all the logic to manage the queues
type Pool struct {
	numRoutines int
	jobs chan Job
	results chan Result
	done chan bool
	completed bool
}

// NewManager returns a new manager structure ready to be used.
func NewPool(numRoutines int ) *Pool {

	fmt.Println("Creating a new Pool")
	r := &Pool{numRoutines: numRoutines}
	r.jobs = make(chan Job, numRoutines)
	r.results = make(chan Result, numRoutines)

	return r
}

func (m *Pool) Start(resources []interface{}, proFunc ProcessorFunc, resFunc ResultProcessorFunc) {

	fmt.Println("Worker pool starting")

	startTime := time.Now()
	go m.allocate(resources)
	m.done = make(chan bool)
	go m.collect(resFunc)
	go m.workerPool(proFunc)
	<-m.done
	endTime := time.Now()
	diff := endTime.Sub(startTime)
	fmt.Println("total time taken: [%f] seconds", diff.Seconds())

}

// allocate allocates jobs based on an array of resources to be processed by the worker pool

func (m *Pool) allocate(jobs []interface{}) {
	defer close(m.jobs)
	fmt.Println("Allocating [%d] resources", len(jobs))

	for i, v := range jobs {

		job := Job{id: i, resource: v}
		m.jobs <- job
	}

	fmt.Println("Done Allocating")
}

// work performs the actual work by calling the processor and passing in the Job as reference obtained
// from iterating over the "Jobs" channel

func (m *Pool) work(wg *sync.WaitGroup, processor ProcessorFunc) {

	defer wg.Done()
	fmt.Print("goRoutine work starting")

	for job := range m.jobs {
		fmt.Println("working on Job ID [%d]", job.id)
		output := Result{job, processor(job.resource)}
		m.results <- output
		fmt.Println("done with Job ID [%d]", job.id)
	}

	fmt.Print("goRoutine work done.")

}

// workerPool creates or spawns new "work" goRoutines to process the "Jobs" channel

func (m *Pool) workerPool(processor ProcessorFunc) {

	defer close(m.results)
	fmt.Println("Worker Pool spawning new goRoutines, total: [%d]", m.numRoutines)
	var wg sync.WaitGroup
	for i := 0; i < m.numRoutines; i++ {
		wg.Add(1)
		go m.work(&wg, processor)
		fmt.Println("Spawned work goRoutine [%d]", i)
	}
	fmt.Println("Worker Pool done spawning work goRoutines")
	wg.Wait()
	fmt.Println("all work goroutines done processing")

}

// Collect post processes the channel "Results" and calls the ResultProcessorFunc passed in as reference
// for further processing.

func (m *Pool) collect(proc ResultProcessorFunc) {
	fmt.Print("goRoutine collect starting")

	for result := range m.results {
		outcome := proc(result)
		fmt.Println("Job with id: [%d] completed, outcome: %s", result.Job.id, outcome)

	}
	fmt.Println("goRoutine collect done, setting channel done as completed")
	m.done <- true
	m.completed = true
}


// IsCompleted utility method to check if all work has done from an outside caller.
func (m *Pool) IsCompleted() bool {
	return m.completed
}






