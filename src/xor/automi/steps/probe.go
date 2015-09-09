package steps

import (
	"fmt"
	"xor/automi/api"
)

type ProbeFunc func(api.Item) api.Item

// Probe is a step designed for testing and inspecting data flow
// It captures and emmit data to provided function probes.
type Probe struct {
	Name    string
	Input   api.Step
	Examine ProbeFunc

	items   chan api.Item
	errs    chan api.ErrorItem
	channel *api.DefaultChannel
}

func (p *Probe) init() error {
	// validation
	if p.Name == "" {
		return fmt.Errorf("Step Probe missing name identifier")
	}

	if p.Input == nil {
		return fmt.Errorf("Probe [%s] missing input attribute")
	}

	p.items = make(chan api.Item)
	p.errs = make(chan api.ErrorItem)
	p.channel = api.NewChannel(p.items, p.errs)
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

	input := p.GetInput().GetChannel()

	go func() {
		defer func() {
			close(p.items)
			close(p.errs)
		}()

		for item := range input.Items() {
			if p.Examine != nil {
				p.items <- p.Examine(item)
			}
		}
	}()
	return nil
}
