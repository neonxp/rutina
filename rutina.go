package rutina

import (
	"context"
	"errors"
	"os"
	"os/signal"
	"sync"
	"sync/atomic"
	"syscall"
	"time"
)

var (
	ErrRunLimit        = errors.New("rutina run limit")
	ErrTimeoutOrKilled = errors.New("rutina timeouted or killed")
	ErrProcessNotFound = errors.New("process not found")
	ErrShutdown        = errors.New("shutdown")
)

type logger func(format string, v ...interface{})

var nopLogger = func(format string, v ...interface{}) {}

// Rutina is routine manager
type Rutina struct {
	ctx               context.Context // State of application (started/stopped)
	Cancel            func()          // Cancel func that stops all routines
	wg                sync.WaitGroup  // WaitGroup that wait all routines to complete
	onceErr           sync.Once       // Flag that prevents overwrite first error that shutdowns all routines
	onceWait          sync.Once       // Flag that prevents wait already waited rutina
	err               error           // First error that shutdowns all routines
	logger            logger          // Optional logger
	counter           *uint64         // Optional counter that names routines with increment ids for debug purposes at logger
	errCh             chan error      // Optional channel for errors when RestartIfError and DoNothingIfError
	autoListenSignals []os.Signal     // Optional listening os signals, default disabled
	processes         map[uint64]*process
	mu                sync.Mutex
}

// New instance with builtin context
func New(opts ...Options) *Rutina {
	if opts == nil {
		opts = []Options{}
	}
	options := composeOptions(opts)
	ctx, cancel := context.WithCancel(options.ParentContext)
	var counter uint64
	return &Rutina{
		ctx:               ctx,
		Cancel:            cancel,
		wg:                sync.WaitGroup{},
		onceErr:           sync.Once{},
		onceWait:          sync.Once{},
		err:               nil,
		logger:            options.Logger,
		counter:           &counter,
		errCh:             options.Errors,
		autoListenSignals: options.ListenOsSignals,
		processes:         map[uint64]*process{},
		mu:                sync.Mutex{},
	}
}

// Go routine
func (r *Rutina) Go(doer func(ctx context.Context) error, opts ...RunOptions) uint64 {
	options := composeRunOptions(opts)
	// Check that context is not canceled yet
	if r.ctx.Err() != nil {
		return 0
	}

	r.mu.Lock()
	id := atomic.AddUint64(r.counter, 1)
	process := process{
		id:           id,
		doer:         doer,
		onDone:       options.OnDone,
		onError:      options.OnError,
		restartLimit: options.MaxCount,
		restartCount: 0,
		timeout:      options.Timeout,
	}
	r.processes[id] = &process
	r.mu.Unlock()

	r.wg.Add(1)
	go func() {
		defer r.wg.Done()
		if err := process.run(r.ctx, r.errCh, r.logger); err != nil {
			if err != ErrShutdown {
				r.onceErr.Do(func() {
					r.err = err
				})
			}
			r.Cancel()
		}
		r.mu.Lock()
		defer r.mu.Unlock()
		delete(r.processes, process.id)
		r.logger("completed #%d", process.id)
	}()
	return id
}

func (r *Rutina) Processes() []uint64 {
	var procesess []uint64
	for id, _ := range r.processes {
		procesess = append(procesess, id)
	}
	return procesess
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
		r.logger("starting OS signals listener")
		select {
		case s := <-sig:
			r.logger("stopping by OS signal (%v)", s)
			r.Cancel()
		case <-r.ctx.Done():
		}
	}()
}

// Wait all routines and returns first error or nil if all routines completes without errors
func (r *Rutina) Wait() error {
	if len(r.autoListenSignals) > 0 {
		r.ListenOsSignals(r.autoListenSignals...)
	}
	r.onceWait.Do(func() {
		r.wg.Wait()
		if r.errCh != nil {
			close(r.errCh)
		}
	})
	return r.err
}

// Kill process by id
func (r *Rutina) Kill(id uint64) error {
	p, ok := r.processes[id]
	if !ok {
		return ErrProcessNotFound
	}
	if p.cancel != nil {
		p.cancel()
	}
	return nil
}

type process struct {
	id           uint64
	doer         func(ctx context.Context) error
	cancel       func()
	onDone       Policy
	onError      Policy
	restartLimit *int
	restartCount int
	timeout      *time.Duration
}

func (p *process) run(pctx context.Context, errCh chan error, logger logger) error {
	var ctx context.Context
	if p.timeout != nil {
		ctx, p.cancel = context.WithTimeout(pctx, *p.timeout)
		defer p.cancel()
	} else {
		ctx, p.cancel = context.WithCancel(pctx)
	}
	for {
		logger("starting process #%d", p.id)
		p.restartCount++
		currentAction := p.onDone
		err := p.doer(ctx)
		if err != nil {
			if p.onError == Shutdown {
				return err
			}
			currentAction = p.onError
			logger("error on process #%d: %s", p.id, err)
			if errCh != nil {
				errCh <- err
			}
		}
		switch currentAction {
		case DoNothing:
			return nil
		case Shutdown:
			return ErrShutdown
		case Restart:
			if ctx.Err() != nil {
				if p.onError == Shutdown {
					return ErrTimeoutOrKilled
				} else {
					if errCh != nil {
						errCh <- ErrTimeoutOrKilled
					}
					return nil
				}
			}
			if p.restartLimit == nil || p.restartCount > *p.restartLimit {
				logger("run count limit process #%d", p.id)
				if p.onError == Shutdown {
					return ErrRunLimit
				} else {
					if errCh != nil {
						errCh <- ErrRunLimit
					}
					return ErrRunLimit
				}
			}
			logger("restarting process #%d", p.id)
		}
	}
}
