package route

import (
	"fmt"
	"sync"

	"github.com/vladimirvivien/automi/api"
)

// ErrCollector aggregates errors from different components
// into a single stream of errors that can then be logged into a source.
type ErrCollector struct {
	Name  string
	Input []<-chan api.ProcError

	output chan api.ProcError
}

func (e *ErrCollector) Init() error {
	if e.Name == "" {
		return api.ProcError{Err: fmt.Errorf("Missing Name attribute")}
	}

	if e.Input == nil {
		return api.ProcError{
			ProcName: e.Name,
			Err:      fmt.Errorf("Missing input attribute"),
		}
	}

	e.output = make(chan api.ProcError)

	return nil
}

func (e *ErrCollector) Uninit() error {
	return nil
}

func (e *ErrCollector) GetName() string {
	return e.Name
}

func (e *ErrCollector) GetOutput() <-chan api.ProcError {
	return e.output
}

func (e *ErrCollector) Exec() (err error) {
	if len(e.Input) == 0 {
		return
	}

	var barrier sync.WaitGroup
	barrier.Add(len(e.Input))
	for _, errCh := range e.Input {
		go func() {
			defer barrier.Done()
			e.merge(errCh)
		}()
	}

	go func() {
		defer func() {
			fmt.Println("Done, closing output")
			close(e.output)
		}()
		barrier.Wait()
	}()

	return nil
}

func (e *ErrCollector) merge(errCh <-chan api.ProcError) {
	fmt.Println("Merging errors")
	for err := range errCh {
		fmt.Println("Merged:", err)
		e.output <- err
	}
	fmt.Println("Done merging")
}
