package api

import (
	"fmt"
	"sync"

	"golang.org/x/net/context"

	"github.com/Sirupsen/logrus"
	autoctx "github.com/vladimirvivien/automi/context"
)

// UnaryOp represents a unary operation (i.e. transformation, etc)
type UnaryOp struct {
	ctx         context.Context
	op          UnOperation
	concurrency int
	input       <-chan interface{}
	output      chan interface{}
	log         *logrus.Entry
	cancelled   bool
	mutex       sync.RWMutex
}

func NewUnaryOp(ctx context.Context) *UnaryOp {
	// extract logger
	log, ok := autoctx.GetLogEntry(ctx)
	if !ok {
		log = logrus.WithField("Component", "UnaryOp")
		log.Error("Logger not found in context")
	}

	o := new(UnaryOp)
	o.ctx = ctx
	o.log = log.WithFields(logrus.Fields{
		"Component": "UnaryOp",
		"Type":      fmt.Sprintf("%T", o),
	})

	o.concurrency = 1
	o.output = make(chan interface{}, 1024)

	o.log.Infof("Component [%s] initialized", "UnaryOp")
	return o
}

func (o *UnaryOp) SetOperation(op UnOperation) {
	o.op = op
}

func (o *UnaryOp) SetConcurrency(concurr int) {
	o.concurrency = concurr
	if o.concurrency < 1 {
		o.concurrency = 1
	}
}

func (o *UnaryOp) SetInput(in <-chan interface{}) {
	o.input = in
}

func (o *UnaryOp) GetOutput() <-chan interface{} {
	return o.output
}

func (o *UnaryOp) Exec() (err error) {
	if o.input == nil {
		err = fmt.Errorf("No input channel found")
		return
	}

	// validate p
	if o.concurrency < 1 {
		o.concurrency = 1
	}

	o.log.Info("Execution started for component UnaryOp")

	go func() {
		defer func() {
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
				o.log.Infof("Component [%s] cancelling...")
				return
			}
		case <-o.ctx.Done():
			o.log.Info("UnaryOp done.")
			return
		}
	}()
	return nil
}

func (o *UnaryOp) doProc(ctx context.Context) {
	if o.op == nil {
		o.log.Error("No operation defined for UnaryOp")
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

			result := o.op.Apply(exeCtx, item)

			switch val := result.(type) {
			case nil:
				continue
			case error, ProcError:
				o.log.Error(val)
				continue
			default:
				o.output <- val
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
