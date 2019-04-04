package rutina

import (
	"context"
	"log"
	"os"
	"os/signal"
	"sync"
	"sync/atomic"
	"syscall"
)

//Rutina is routine manager
type Rutina struct {
	ctx     context.Context // State of application (started/stopped)
	Cancel  func()          // Cancel func that stops all routines
	wg      sync.WaitGroup  // WaitGroup that wait all routines to complete
	o       sync.Once       // Flag that prevents overwrite first error that shutdowns all routines
	err     error           // First error that shutdowns all routines
	logger  *log.Logger     // Optional logger
	counter *uint64         // Optional counter that names routines with increment ids for debug purposes at logger
	errCh   chan error      // Optional channel for errors when RestartIfFail and DoNothingIfFail
}

// New instance with builtin context
func New(mixins ...Mixin) *Rutina {
	ctx, cancel := context.WithCancel(context.Background())
	var counter uint64 = 0
	r := &Rutina{ctx: ctx, Cancel: cancel, counter: &counter, errCh: nil}
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
	onDone := ShutdownIfDone
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
		r.log("starting #%d", id)
		if err := doer(r.ctx); err != nil {
			// errors history
			if r.errCh != nil {
				r.errCh <- err
			}
			// region routine failed
			r.log("error at #%d : %v", id, err)
			switch onFail {
			case ShutdownIfFail:
				r.log("stopping #%d", id)
				// Save error only if shutdown all routines
				r.o.Do(func() {
					r.err = err
				})
				r.Cancel()
			case RestartIfFail:
				r.log("restarting #%d", id)
				r.Go(doer, opts...)
			case DoNothingIfFail:
				r.log("stopping #%d", id)
			}
			// endregion
		} else {
			// region routine successfully done
			switch onDone {
			case ShutdownIfDone:
				r.log("stopping #%d with shutdown", id)
				r.Cancel()
			case RestartIfDone:
				r.log("restarting #%d", id)
				r.Go(doer, opts...)
			case DoNothingIfDone:
				r.log("stopping #%d", id)
			}
			// endregion
		}

	}()
}

// OS signals handler
func (r *Rutina) ListenOsSignals(signals ...os.Signal) {
	if len(signals) == 0 {
		signals = []os.Signal{syscall.SIGINT, syscall.SIGTERM}
	}
	go func() {
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, signals...)
		r.log("starting OS signals listener")
		select {
		case s := <-sig:
			r.log("stopping by OS signal (%v)", s)
			r.Cancel()
		case <-r.ctx.Done():
		}
	}()
}

// Wait all routines and returns first error or nil if all routines completes without errors
func (r *Rutina) Wait() error {
	r.wg.Wait()
	return r.err
}

// Log if can
func (r *Rutina) log(format string, args ...interface{}) {
	if r.logger != nil {
		r.logger.Printf(format, args...)
	}
}
