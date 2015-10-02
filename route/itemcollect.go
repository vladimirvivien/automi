package route

import (
	"fmt"
	"sync"

	"golang.org/x/net/context"

	"github.com/Sirupsen/logrus"
	"github.com/vladimirvivien/automi/api"
	autoctx "github.com/vladimirvivien/automi/context"
)

// ItemCollector aggregates data itesm from different source components
// into a single stream of data that can then be sinked into another component.
type ItemCollector struct {
	Name string

	inputs []<-chan interface{}
	output chan interface{}
	log    *logrus.Entry
}

func (c *ItemCollector) Init(ctx context.Context) error {
	// extract logger
	log, ok := autoctx.GetLogEntry(ctx)
	if !ok {
		log = logrus.WithField("Proc", "ItemCollector")
		log.Error("Logger not set for component")
	}

	c.log = log.WithFields(logrus.Fields{
		"Component": c.Name,
		"Type":      fmt.Sprintf("%T", c),
	})

	if c.Name == "" {
		return api.ProcError{Err: fmt.Errorf("Missing Name attribute")}
	}

	if c.inputs == nil {
		return api.ProcError{
			ProcName: c.Name,
			Err:      fmt.Errorf("Missing Inputs attribute"),
		}
	}

	c.output = make(chan interface{})
	c.log.Info("Component initialized")
	return nil
}

func (e *ItemCollector) Uninit(ctx context.Context) error {
	return nil
}

func (c *ItemCollector) GetName() string {
	return c.Name
}

func (c *ItemCollector) SetInputs(ins []<-chan interface{}) {
	c.inputs = ins
}

func (c *ItemCollector) GetOutput() <-chan interface{} {
	return c.output
}

func (c *ItemCollector) Exec(ctx context.Context) (err error) {
	if len(c.inputs) == 0 {
		return api.ProcError{
			ProcName: c.Name,
			Err:      fmt.Errorf("No inputs set for component"),
		}
	}

	var barrier sync.WaitGroup
	barrier.Add(len(c.inputs))

	for _, itemChan := range c.inputs {
		go func(ic <-chan interface{}) {
			c.merge(&barrier, ic)
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
