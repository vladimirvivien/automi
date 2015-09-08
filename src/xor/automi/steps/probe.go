package steps

import (
	"fmt"
	"xor/automi/api"
)

type probeChan chan api.Item

func (p probeChan) Extract() <-chan api.Item {
	return p
}

type ProbeFunc func(api.Item) api.Item

// Probe is a step designed for testing and inspecting data flow
// It captures and emmit data to provided function probes.
type Probe struct {
	Name    string
	Input   api.Step
	Examine ProbeFunc
	channel probeChan
}

func (p *Probe) init() error {
	// validation
	if p.Name == "" {
		return fmt.Errorf("Step Probe missing name identifier")
	}

	if p.Input == nil {
		return fmt.Errorf("Probe [%s] missing input attribute")
	}

	p.channel = make(chan api.Item)
	return nil
}

func (p *Probe) GetName() string {
	return p.Name
}

func (p *Probe) GetChannel() api.Channel {
	return p.channel
}

func (p *Probe) GetInput() api.Step {
	return p.Input
}

func (p *Probe) Do() error {
	if err := p.init(); err != nil {
		return err
	}

	input := p.GetInput()
	if input == nil {
		return fmt.Errorf("Probe [%s] has no input specified", p.GetName())
	}
	go func() {
		defer func() {
			close(p.channel)
		}()

		for item := range p.GetInput().GetChannel().Extract() {
			if p.Examine != nil {
				p.channel <- p.Examine(item)
			}
		}
	}()
	return nil
}
