package rutina

import (
	"context"
	"log"
	"os"
	"os/signal"
	"sync"
	"sync/atomic"
)

//Rutina is routine manager
type Rutina struct {
	ctx           context.Context
	Cancel        func()
	wg            sync.WaitGroup
	o             sync.Once
	err           error
	logger        *log.Logger
	counter       *uint64
	cancelByError bool
}

// New instance with builtin context
func New(opts ...Option) (*Rutina, context.Context) {
	ctx, cancel := context.WithCancel(context.Background())
	var counter uint64 = 0
	r := &Rutina{ctx: ctx, Cancel: cancel, counter: &counter, cancelByError: false}
	return r.WithOptions(opts...), ctx
}

func (r *Rutina) WithOptions(opts ...Option) *Rutina {
	nr := *r
	for _, o := range opts {
		o.apply(&nr)
	}
	return &nr
}

// Go routine
func (r *Rutina) Go(doer func(ctx context.Context) error) {
	r.wg.Add(1)
	go func() {
		id := atomic.AddUint64(r.counter, 1)
		defer func() {
			if r.logger != nil {
				r.logger.Printf("stopping #%d", id)
			}
			r.wg.Done()
			if !r.cancelByError {
				r.Cancel()
			}
		}()
		if r.logger != nil {
			r.logger.Printf("starting #%d", id)
		}
		if err := doer(r.ctx); err != nil {
			if r.logger != nil {
				r.logger.Printf("error at #%d : %v", id, err)
			}
			r.o.Do(func() {
				r.err = err
			})
			if r.cancelByError {
				r.Cancel()
			}
		}
	}()
}

// OS signals handler
func (r *Rutina) ListenOsSignals() {
	r.Go(func(ctx context.Context) error {
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, os.Interrupt, os.Kill)
		select {
		case s := <-sig:
			if r.logger != nil {
				r.logger.Printf("stopping by OS signal (%v)", s)
			}
			if r.cancelByError {
				r.Cancel()
			}
		case <-ctx.Done():
		}
		return nil
	})
}

// Wait all routines and returns first error or nil if all routines completes without errors
func (r *Rutina) Wait() error {
	r.wg.Wait()
	return r.err
}
