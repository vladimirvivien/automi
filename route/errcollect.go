package route

import (
	"fmt"
	"sync"

	"github.com/vladimirvivien/automi/api"
)

// ItemCollector aggregates data itesm from different source components
// into a single stream of data that can then be sinked into another component.
type ItemCollector struct {
	Name   string
	Inputs []<-chan interface{}

	output chan interface{}
}

func (c *ItemCollector) Init() error {
	if c.Name == "" {
		return api.ProcError{Err: fmt.Errorf("Missing Name attribute")}
	}

	if c.Inputs == nil {
		return api.ProcError{
			ProcName: c.Name,
			Err:      fmt.Errorf("Missing Inputs attribute"),
		}
	}

	c.output = make(chan interface{})

	return nil
}

func (e *ItemCollector) Uninit() error {
	return nil
}

func (c *ItemCollector) GetName() string {
	return c.Name
}

func (c *ItemCollector) GetOutput() <-chan interface{} {
	return c.output
}

func (c *ItemCollector) Exec() (err error) {
	if len(c.Inputs) == 0 {
		return
	}

	var barrier sync.WaitGroup
	barrier.Add(len(c.Inputs))
	for _, itemChan := range c.Inputs {
		go func(ec <-chan interface{}) {
			c.merge(&barrier, ec)
		}(itemChan)
	}

	go func() {
		defer func() {
			close(c.output)
		}()
		barrier.Wait()
	}()

	return nil
}

func (c *ItemCollector) merge(wg *sync.WaitGroup, ch <-chan interface{}) {
	for item := range ch {
		c.output <- item
	}
	wg.Done()
}
