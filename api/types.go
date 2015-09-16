package api

import "fmt"

type Process interface {
	GetName() string
	Exec() error
	Init() error
	Uninit() error
}

type ErrProducer interface {
	GetErrors() <-chan ProcError
}

type Processor interface {
	Process
	GetInput() <-chan interface{}
	GetOuput() <-chan interface{}
}

type Source interface {
	Process
	GetOutput() <-chan interface{}
}

type Sink interface {
	Process
	GetInput() <-chan interface{}
}

type Collector interface {
	Process
	GetInputs() []<-chan interface{}
	GetOutput() <-chan interface{}
}

type Emitter interface {
	Process
	GetInput() <-chan interface{}
	GetOutputs() []<-chan interface{}
}

type ProcError struct {
	Err      error
	ProcName string
}

func (e ProcError) Error() string {
	return fmt.Sprintf("[%s] %v", e.ProcName, e.Err)
}
