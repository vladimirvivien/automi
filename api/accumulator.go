package api

import (
	"fmt"
	"sync"

	"github.com/Sirupsen/logrus"
	autoctx "github.com/vladimirvivien/automi/context"

	"golang.org/x/net/context"
)

type Accumulator struct {
	ctx         context.Context
	op          UnOperation
	concurrency int
	input       <-chan interface{}
	output      chan interface{}
	log         *logrus.Entry
	cancelled   bool
	mutex       sync.RWMutex
	state       []interface{}
}

func NewAccumulator(ctx context.Context) *Accumulator {
	// extract logger
	log, ok := autoctx.GetLogEntry(ctx)
	if !ok {
		log = logrus.WithField("Component", "Accumulator")
		log.Error("Logger not found in context")
	}

	a := new(Accumulator)
	a.ctx = ctx
	a.log = log.WithFields(logrus.Fields{
		"Component": "Accumulator",
		"Type":      fmt.Sprintf("%T", a),
	})

	a.concurrency = 1
	a.output = make(chan interface{}, 1024)

	a.log.Infof("Component [%s] initialized", "Accumulator")
	return a
}

func (a *Accumulator) SetConcurrency(concurr int) {
	a.concurrency = concurr
	if a.concurrency < 1 {
		a.concurrency = 1
	}
}

func (a *Accumulator) SetInput(in <-chan interface{}) {
	a.input = in
}

func (a *Accumulator) GetOutput() <-chan interface{} {
	return a.output
}

func (a *Accumulator) Exec() (err error) {
	if a.input == nil {
		err = fmt.Errorf("No input channel found")
		return
	}

	// validate p
	if a.concurrency < 1 {
		a.concurrency = 1
	}

	a.log.Info("Execution started")

	go func() {
		defer func() {
			close(a.output)
			a.log.Info("Shuttingdown...")
		}()

		var barrier sync.WaitGroup
		wgDelta := a.concurrency
		barrier.Add(wgDelta)

		for i := 0; i < a.concurrency; i++ { // workers
			go func(wg *sync.WaitGroup) {
				defer wg.Done()
				a.doProc(a.ctx)
			}(&barrier)
		}

		wait := make(chan struct{})
		go func() {
			defer close(wait)
			barrier.Wait()
		}()

		select {
		case <-wait:
			if a.cancelled {
				a.log.Infof("Component cancelling...")
				return
			}
		case <-a.ctx.Done():
			a.log.Info("Component done")
			return
		}
	}()
	return nil
}

func (a *Accumulator) doProc(ctx context.Context) {
	if a.op == nil {
		a.log.Error("No operation defined for Accumulator")
		return
	}
	exeCtx, cancel := context.WithCancel(ctx)

	for {
		select {
		// process incoming item
		case item, opened := <-a.input:
			if !opened {
				return
			}

			result := a.op.Apply(exeCtx, item)

			switch val := result.(type) {
			case nil:
				continue
			case error, ProcError:
				a.log.Error(val)
				continue
			default:
				a.output <- val
			}

		// is cancelling
		case <-ctx.Done():
			a.log.Infoln("Cancelling....")
			a.mutex.Lock()
			cancel()
			a.cancelled = true
			a.mutex.Unlock()
			return
		}
	}
}
