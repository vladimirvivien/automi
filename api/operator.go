package api

import (
	"fmt"
	"sync"

	"golang.org/x/net/context"

	"github.com/Sirupsen/logrus"
	autoctx "github.com/vladimirvivien/automi/context"
)

// Operator lauches an operation.
type Operator struct {
	op          Operation
	concurrency int
	input       <-chan interface{}
	output      chan interface{}
	log         *logrus.Entry
	cancelled   bool
	mutex       sync.RWMutex
}

func NewOperator(ctx context.Context) *Operator {
	// extract logger
	log, ok := autoctx.GetLogEntry(ctx)
	if !ok {
		log = logrus.WithField("Component", "Operator")
		log.Error("Logger not found in context")
	}

	o := new(Operator)

	o.log = log.WithFields(logrus.Fields{
		"Component": "Operator",
		"Type":      fmt.Sprintf("%T", o),
	})

	o.concurrency = 1
	o.output = make(chan interface{}, 1024)

	o.log.Infof("Component [%s] initialized", "Operator")
	return o
}

func (o *Operator) SetOperation(op Operation) {
	o.op = op
}

func (o *Operator) SetConcurrency(concurr int) {
	o.concurrency = concurr
	if o.concurrency < 1 {
		o.concurrency = 1
	}
}

func (o *Operator) SetInput(in <-chan interface{}) {
	o.input = in
}

func (o *Operator) GetOutput() <-chan interface{} {
	return o.output
}

func (o *Operator) Exec(ctx context.Context) (err error) {
	if o.input == nil {
		err = fmt.Errorf("No input channel found")
		return
	}

	// validate p
	if o.concurrency < 1 {
		o.concurrency = 1
	}

	o.log.Info("Execution started for component Operator")

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
				o.doProc(ctx)
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
		case <-ctx.Done():
			o.log.Info("Operator done.")
			return
		}
	}()
	return nil
}

func (o *Operator) doProc(ctx context.Context) {
	if o.op == nil {
		o.log.Error("No operation defined for Operator")
		return
	}
	exeCtx, cancel := context.WithCancel(ctx)

	o.log.Info("About to start processing....")
	for {
		o.log.Info("Processing data ...")
		select {
		// process incoming item
		case item, opened := <-o.input:
			o.log.Info("About to process ", item)
			if !opened {
				o.log.Info("Channel closed, bailing")
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
	o.log.Info("Nothing to do, exiting")
}
