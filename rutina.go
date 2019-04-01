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
func New(mixins ...Mixin) *Rutina {
	ctx, cancel := context.WithCancel(context.Background())
	var counter uint64 = 0
	r := &Rutina{ctx: ctx, Cancel: cancel, counter: &counter, cancelByError: false}
	return r.With(mixins...)
}

func (r *Rutina) With(mixins ...Mixin) *Rutina {
	nr := *r
	for _, m := range mixins {
		m.apply(&nr)
	}
	return &nr
}

// Go routine
func (r *Rutina) Go(doer func(ctx context.Context) error, opts ...Options) {
	onFail := ShutdownIfFail
	for _, o := range opts {
		switch o {
		case ShutdownIfFail:
			onFail = ShutdownIfFail
		case RestartIfFail:
			onFail = RestartIfFail
		case DoNothingIfFail:
			onFail = DoNothingIfFail
		}
	}
	onDone := DoNothingIfDone
	for _, o := range opts {
		switch o {
		case ShutdownIfDone:
			onDone = ShutdownIfDone
		case RestartIfDone:
			onDone = RestartIfDone
		case DoNothingIfDone:
			onDone = DoNothingIfDone
		}
	}

	r.wg.Add(1)
	go func() {
		defer r.wg.Done()
		// Check that context is not canceled yet
		if r.ctx.Err() != nil {
			return
		}
		id := atomic.AddUint64(r.counter, 1)
		if r.logger != nil {
			r.logger.Printf("starting #%d", id)
		}
		if err := doer(r.ctx); err != nil {
			if r.logger != nil {
				r.logger.Printf("error at #%d : %v", id, err)
			}
			switch onFail {
			case ShutdownIfFail:
				if r.logger != nil {
					r.logger.Printf("stopping #%d", id)
				}
				// Save error only if shutdown all routines
				r.o.Do(func() {
					r.err = err
				})
				r.Cancel()
			case RestartIfFail:
				// TODO maybe store errors on restart?
				if r.logger != nil {
					r.logger.Printf("restarting #%d", id)
				}
				r.Go(doer, opts...)
			case DoNothingIfFail:
				// TODO maybe store errors on nothing to do?
				if r.logger != nil {
					r.logger.Printf("stopping #%d", id)
				}
			}
		} else {
			switch onDone {
			case ShutdownIfDone:
				if r.logger != nil {
					r.logger.Printf("stopping #%d with shutdown", id)
				}
				r.Cancel()
			case RestartIfDone:
				if r.logger != nil {
					r.logger.Printf("restarting #%d", id)
				}
				r.Go(doer, opts...)
			case DoNothingIfDone:
				if r.logger != nil {
					r.logger.Printf("stopping #%d", id)
				}
			}
		}

	}()
}

// OS signals handler
func (r *Rutina) ListenOsSignals(signals ...os.Signal) {
	if len(signals) == 0 {
		signals = []os.Signal{os.Kill, os.Interrupt}
	}
	r.Go(func(ctx context.Context) error {
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, signals...)
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
	}, ShutdownIfDone)
}

// Wait all routines and returns first error or nil if all routines completes without errors
func (r *Rutina) Wait() error {
	r.wg.Wait()
	return r.err
}
