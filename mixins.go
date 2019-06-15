package rutina

import (
	"context"
	"log"
	"os"
)

// Mixin interface
type Mixin interface {
	apply(*Rutina)
}

// MixinContext propagates user defined context to rutina
type MixinContext struct {
	Context context.Context
}

// WithContext propagates user defined context to rutina
func WithContext(context context.Context) *MixinContext {
	return &MixinContext{Context: context}
}

func (o MixinContext) apply(r *Rutina) {
	ctx, cancel := context.WithCancel(o.Context)
	r.ctx = ctx
	r.Cancel = cancel
}

// MixinLogger adds logger to rutina
type MixinLogger struct {
	Logger *log.Logger
}

// WithLogger adds custom logger to rutina
func WithLogger(logger *log.Logger) *MixinLogger {
	return &MixinLogger{Logger: logger}
}

// WithStdLogger adds standard logger to rutina
func WithStdLogger() *MixinLogger {
	return &MixinLogger{Logger: log.New(os.Stdout, "[rutina]", log.LstdFlags)}
}

func (o MixinLogger) apply(r *Rutina) {
	r.logger = o.Logger
}

// MixinErrChan turns on errors channel on rutina
type MixinErrChan struct {
}

// WithErrChan turns on errors channel on rutina
func WithErrChan() *MixinErrChan {
	return &MixinErrChan{}
}

func (o MixinErrChan) apply(r *Rutina) {
	r.errCh = make(chan error, 1)
}

type LifecycleListener func(event Event, routineID int)

type LifecycleMixin struct {
	Listener LifecycleListener
}

func (l LifecycleMixin) apply(r *Rutina) {
	r.lifecycleListener = l.Listener
}

func WithLifecycleListener(listener LifecycleListener) *LifecycleMixin {
	return &LifecycleMixin{Listener: listener}
}
