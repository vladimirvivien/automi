package api

import "errors"

var (
	ErrSourceInputUndefined     = errors.New("source input undefined")
	ErrSinkDestinationUndefined = errors.New("sink destination undefined")
	ErrSourceEmpty              = errors.New("source is empty")
	ErrStreamEmpty              = errors.New("stream is empty")
	ErrSinkEmpty                = errors.New("sink is empty")
	ErrInputChannelUndefined    = errors.New("undefined input channel")
	ErrSourceUndefined          = errors.New("source undefined")
)

// // StreamError is used to signal runtime stream error
// type StreamError[T any] struct {
// 	err  string        // Error message
// 	item StreamItem[T] // Item that caused error
// }

// // Error returns a string value for StreamError
// func (e StreamError[T]) Error() string {
// 	return e.err
// }

// // Item returns the StreamItem associated with the error
// func (e StreamError[T]) Item() StreamItem[T] {
// 	return e.item
// }

// // Error returns a StreamError
// func Error(msg string) StreamError[any] {
// 	return StreamError[any]{err: msg}
// }

// // ErrorWithItem returns a StreamError with provided StreamItem
// func ErrorWithItem[T any](msg string, item StreamItem[T]) StreamError[T] {
// 	return StreamError[T]{err: msg, item: item}
// }

// // CancelStreamError signals that all stream activities should stop
// // and the streaming should gracefully end
// type CancelStreamError[T any] StreamError[T]

// // Error returns a string value for CancelStreamError
// func (e CancelStreamError[T]) Error() string {
// 	return e.err
// }

// // CancellationError returns a CancelStreamError
// func CancellationError(msg string) CancelStreamError[any] {
// 	return CancelStreamError[any](Error(msg))
// }
