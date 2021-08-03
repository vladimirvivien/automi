package emitters

import (
	"context"
	"errors"
	"reflect"

	"github.com/vladimirvivien/automi/api"
	autoctx "github.com/vladimirvivien/automi/api/context"
	"github.com/vladimirvivien/automi/util"
)

// ChanEmitter is an emitter that takes in a channel and
// sets it up as the source of the emitter .
type ChanEmitter struct {
	channel interface{}
	output  chan interface{}
	logf    api.LogFunc
}

// Chan is a construction function that returns a new ChanEmitter
func Chan(channel interface{}) *ChanEmitter {
	return &ChanEmitter{
		channel: channel,
		output:  make(chan interface{}, 1024),
	}
}

//GetOutput returns the output channel of this source node
func (c *ChanEmitter) GetOutput() <-chan interface{} {
	return c.output
}

// Open opens the source node to start streaming data on its channel
func (c *ChanEmitter) Open(ctx context.Context) error {
	// ensure channel param is a chan type
	chanType := reflect.TypeOf(c.channel)
	if chanType.Kind() != reflect.Chan {
		return errors.New("ChanEmitter requires channel")
	}
	c.logf = autoctx.GetLogFunc(ctx)
	util.Logfn(c.logf, "Opening channel emitter")
	chanVal := reflect.ValueOf(c.channel)

	if !chanVal.IsValid() {
		return errors.New("invalid channel for ChanEmitter")
	}

	go func() {
		exeCtx, cancel := context.WithCancel(ctx)
		defer func() {
			util.Logfn(c.logf, "Slice emitter closing")
			cancel()
			close(c.output)
		}()

		for {
			val, open := chanVal.Recv()
			if !open {
				return
			}
			select {
			case c.output <- val.Interface():
			case <-exeCtx.Done():
				return
			}
		}
	}()
	return nil
}
