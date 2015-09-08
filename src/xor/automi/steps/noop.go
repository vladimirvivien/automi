package steps

import "xor/automi/api"

type NoOp struct {
	Name    string
	Input   api.Step
	Channel api.Channel
}

func (s *NoOp) GetChannel() api.Channel {
	return s.Channel
}

func (s *NoOp) GetName() string {
	return s.Name
}

func (s *NoOp) GetInput() api.Step {
	return s.Input
}

func (s *NoOp) Do() error {
	return nil
}
