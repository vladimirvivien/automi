package api

type Item interface {
	Get() interface{}
}

type Channel interface {
	Items() <-chan Item
	Errors() <-chan ErrorItem
}

type DefaultChannel struct {
	items <-chan Item
	errs  <-chan ErrorItem
}

func NewChannel(items <-chan Item, errs <-chan ErrorItem) *DefaultChannel {
	return &DefaultChannel{items: items, errs: errs}
}

func (d *DefaultChannel) Items() <-chan Item {
	return d.items
}

func (d *DefaultChannel) Errors() <-chan ErrorItem {
	return d.errs
}

type Step interface {
	GetChannel() Channel
	GetInput() Step
	GetName() string
	Do() error
}

type ErrorItem struct {
	Error    error
	StepName string
	Item     Item
}
