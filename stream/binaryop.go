package stream

import (
	"fmt"
	"sync"

	"golang.org/x/net/context"

	"github.com/Sirupsen/logrus"
	"github.com/vladimirvivien/automi/api"
	autoctx "github.com/vladimirvivien/automi/api/context"
)

// BinaryOp represents a binary operation (i.e. aggregation, reduction, etc)
type BinaryOp struct {
	ctx         context.Context
	op          api.BinOperation
	state       interface{}
	concurrency int
	input       <-chan interface{}
	output      chan interface{}
	log         *logrus.Entry
	cancelled   bool
	mutex       sync.RWMutex
}

func NewBinaryOp(ctx context.Context) *BinaryOp {
	// extract logger
	log, ok := autoctx.GetLogEntry(ctx)
	if !ok {
		log = logrus.WithField("Component", "BinaryOperator")
		log.Error("Logger not found in context")
	}

	o := new(BinaryOp)
	o.ctx = ctx
	o.log = log.WithFields(logrus.Fields{
		"Component": "BinaryOperator",
		"Type":      fmt.Sprintf("%T", o),
	})

	o.concurrency = 1
	o.output = make(chan interface{}, 1024)

	o.log.Infof("Component initialized")
	return o
}

func (o *BinaryOp) SetOperation(op api.BinOperation) {
	o.op = op
}

func (o *BinaryOp) SetInitialState(val interface{}) {
	o.state = val
}

func (o *BinaryOp) SetConcurrency(concurr int) {
	o.concurrency = concurr
	if o.concurrency < 1 {
		o.concurrency = 1
	}
}

func (o *BinaryOp) SetInput(in <-chan interface{}) {
	o.input = in
}

func (o *BinaryOp) GetOutput() <-chan interface{} {
	return o.output
}

func (o *BinaryOp) Exec() (err error) {
	if o.input == nil {
		err = fmt.Errorf("No input channel found")
		return
	}

	// validate p
	if o.concurrency < 1 {
		o.concurrency = 1
	}

	o.log.Info("Execution started")

	go func() {
		defer func() {
			o.output <- o.state
			close(o.output)
			o.log.Info("Shuttingdown component")
		}()

		var barrier sync.WaitGroup
		wgDelta := o.concurrency
		barrier.Add(wgDelta)

		for i := 0; i < o.concurrency; i++ { // workers
			go func(wg *sync.WaitGroup) {
				defer wg.Done()
				o.doProc(o.ctx)
			}(&barrier)
		}

		wait := make(chan struct{})
		go func() {
			defer close(wait)
			barrier.Wait()
		}()

		select {
		case <-wait:
			if o.cancelled {
				o.log.Infof("Component cancelling...")
				return
			}
		case <-o.ctx.Done():
			o.log.Info("BinaryOp done.")
			return
		}
	}()
	return nil
}

func (o *BinaryOp) doProc(ctx context.Context) {
	if o.op == nil {
		o.log.Error("No operation defined for BinaryOp")
		return
	}
	exeCtx, cancel := context.WithCancel(ctx)

	for {
		select {
		// process incoming item
		case item, opened := <-o.input:
			if !opened {
				return
			}

			o.state = o.op.Apply(exeCtx, o.state, item)

			switch val := o.state.(type) {
			case nil:
				continue
			case error, api.ProcError:
				o.log.Error(val)
				continue
			}

		// is cancelling
		case <-ctx.Done():
			o.log.Infoln("Cancelling....")
			o.mutex.Lock()
			cancel()
			o.cancelled = true
			o.mutex.Unlock()
			return
		}
	}
}
