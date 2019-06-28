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
	ctx               context.Context   // State of application (started/stopped)
	Cancel            func()            // Cancel func that stops all routines
	wg                sync.WaitGroup    // WaitGroup that wait all routines to complete
	onceErr           sync.Once         // Flag that prevents overwrite first error that shutdowns all routines
	onceWait          sync.Once         // Flag that prevents wait already waited rutina
	err               error             // First error that shutdowns all routines
	logger            *log.Logger       // Optional logger
	counter           *uint64           // Optional counter that names routines with increment ids for debug purposes at logger
	errCh             chan error        // Optional channel for errors when RestartIfError and DoNothingIfError
	lifecycleListener LifecycleListener // Optional listener for events
	autoListenSignals []os.Signal       // Optional listening os signals, default disabled
}

// New instance with builtin context
func New(mixins ...Mixin) *Rutina {
	ctx, cancel := context.WithCancel(context.Background())
	var counter uint64
	r := &Rutina{ctx: ctx, Cancel: cancel, counter: &counter, errCh: nil}
	return r.With(mixins...)
}

// With applies mixins
func (r *Rutina) With(mixins ...Mixin) *Rutina {
	for _, m := range mixins {
		m.apply(r)
	}
	if r.autoListenSignals != nil {
		r.ListenOsSignals(r.autoListenSignals...)
	}
	return r
}

// Go routine
func (r *Rutina) Go(doer func(ctx context.Context) error, opts ...Options) {
	// Check that context is not canceled yet
	if r.ctx.Err() != nil {
		return
	}
	onFail := ShutdownIfError
	for _, o := range opts {
		switch o {
		case ShutdownIfError:
			onFail = ShutdownIfError
		case RestartIfError:
			onFail = RestartIfError
		case DoNothingIfError:
			onFail = DoNothingIfError
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
		id := atomic.AddUint64(r.counter, 1)
		r.lifecycleEvent(EventRoutineStart, int(id))
		if err := doer(r.ctx); err != nil {
			r.lifecycleEvent(EventRoutineError, int(id))
			r.lifecycleEvent(EventRoutineStop, int(id))
			// errors history
			if r.errCh != nil {
				r.errCh <- err
			}
			// region routine failed
			switch onFail {
			case ShutdownIfError:
				// Save error only if shutdown all routines
				r.onceErr.Do(func() {
					r.err = err
				})
				r.Cancel()
			case RestartIfError:
				r.Go(doer, opts...)
			}
			// endregion
		} else {
			r.lifecycleEvent(EventRoutineComplete, int(id))
			r.lifecycleEvent(EventRoutineStop, int(id))
			// region routine successfully done
			switch onDone {
			case ShutdownIfDone:
				r.Cancel()
			case RestartIfDone:
				r.Go(doer, opts...)
			}
			// endregion
		}
	}()
}

// Errors returns chan for all errors, event if DoNothingIfError or RestartIfError set.
// By default it nil. Use MixinErrChan to turn it on
func (r *Rutina) Errors() <-chan error {
	return r.errCh
}

// ListenOsSignals is simple OS signals handler. By default listen syscall.SIGINT and syscall.SIGTERM
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
	r.onceWait.Do(func() {
		r.wg.Wait()
		r.lifecycleEvent(EventAppStop, 0)
		if r.err == nil {
			r.lifecycleEvent(EventAppComplete, 0)
		} else {
			r.lifecycleEvent(EventAppError, 0)
		}
		if r.errCh != nil {
			close(r.errCh)
		}
	})
	return r.err
}

func (r *Rutina) lifecycleEvent(ev Event, rid int) {
	r.log("Event = %s Routine ID = %d", ev.String(), rid)
	if r.lifecycleListener != nil {
		r.lifecycleListener(ev, rid)
	}
}

// Log if can
func (r *Rutina) log(format string, args ...interface{}) {
	if r.logger != nil {
		r.logger.Printf(format, args...)
	}
}
