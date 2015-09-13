package sup

import (
	"fmt"

	"github.com/vladimirvivien/automi/api"
)

type ProbeFunc func(interface{}) interface{}

// Probe is a processor designed for testing and inspecting data flow
// It captures data in its input channel, apply specified function,
// then outputs the result to its outupt channel
type Probe struct {
	Name    string
	Input   <-chan interface{}
	Output  <-chan interface{}
	Errors  <-chan api.ProcError
	Examine ProbeFunc

	output chan interface{}
	errs   chan api.ProcError
}

func (p *Probe) Init() error {
	// validation
	if p.Name == "" {
		return fmt.Errorf("Step Probe missing name identifier")
	}

	if p.Input == nil {
		return fmt.Errorf("Probe [%s] missing input attribute", p.Name)
	}

	p.output = make(chan interface{})
	p.errs = make(chan api.ProcError)
	return nil
}

func (p *Probe) Uninit() error {
	return nil
}

func (p *Probe) GetName() string {
	return p.Name
}

func (p *Probe) GetInput() <-chan interface{} {
	return p.Input
}

func (p *Probe) GetOutput() <-chan interface{} {
	return p.output
}

func (p *Probe) Exec() error {
	go func() {
		defer func() {
			close(p.output)
			close(p.errs)
		}()

		items := p.GetInput()
		if items == nil {
			panic(fmt.Sprintf("Probe [%s] Input channel nil", p.Name))

		}

		// output data
		for item := range items {
			if p.Examine != nil {
				p.output <- p.Examine(item)
			}
		}
	}()
	return nil
}
