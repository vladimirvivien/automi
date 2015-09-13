package api

type Process interface {
	GetName() string
	Exec() error
	Init() error
	Uninit() error
	GetErrors() <-chan ProcError
}

type Processor interface {
	Process
	GetInput() <-chan interface{}
	GetOuput() <-chan interface{}
}

type Producer interface {
	Process
	GetOutput() <-chan interface{}
}

type Consumer interface {
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
