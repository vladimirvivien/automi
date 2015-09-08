package api

type Item interface {
	Get() interface{}
}

type Channel interface {
	Extract() <-chan Item
}

type Step interface {
	GetChannel() Channel
	GetInput() Step
	GetName() string
	Do() error
}
