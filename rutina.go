package rutina

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

//Rutina is routine manager
type Rutina struct {
	ctx    context.Context
	cancel func()
	wg     sync.WaitGroup
	o      sync.Once
	err    error
}

// New instance with builtin context
func New() (*Rutina, context.Context) {
	return WithContext(context.Background())
}

// WithContext is constructor that takes context from outside
func WithContext(ctx context.Context) (*Rutina, context.Context) {
	ctx, cancel := context.WithCancel(ctx)

	return &Rutina{ctx: ctx, cancel: cancel}, ctx
}

// Go routine
func (r *Rutina) Go(doer func(ctx context.Context) error) {
	r.wg.Add(1)
	go func() {
		defer func() {
			r.wg.Done()
			if r.cancel != nil {
				r.cancel()
			}
		}()
		if err := doer(r.ctx); err != nil {
			r.o.Do(func() {
				r.err = err
			})
		}
	}()
}

// OS signals handler
func (r *Rutina) ListenTermSignals() {
	r.Go(func(ctx context.Context) error {
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, syscall.SIGTERM, syscall.SIGINT)
		select {
		case <-sig:
		case <-ctx.Done():
		}
		return nil
	})
}

// Wait all routines and returns first error or nil if all routines completes without errors
func (r *Rutina) Wait() error {
	r.wg.Wait()
	if r.cancel != nil {
		r.cancel()
	}
	return r.err
}
