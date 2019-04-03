package rutina

import (
	"context"
	"log"
	"os"
)

type Mixin interface {
	apply(*Rutina)
}

type MixinContext struct {
	Context context.Context
}

func WithContext(context context.Context) *MixinContext {
	return &MixinContext{Context: context}
}

func (o MixinContext) apply(r *Rutina) {
	ctx, cancel := context.WithCancel(o.Context)
	r.ctx = ctx
	r.Cancel = cancel
}

type MixinLogger struct {
	Logger *log.Logger
}

func WithLogger(logger *log.Logger) *MixinLogger {
	return &MixinLogger{Logger: logger}
}

func WithStdLogger() *MixinLogger {
	return &MixinLogger{Logger: log.New(os.Stdout, "rutina", log.LstdFlags)}
}

func (o MixinLogger) apply(r *Rutina) {
	r.logger = o.Logger
}

type MixinErrChan struct {
	errCh chan error
}

func WithErrChan(errCh chan error) *MixinErrChan {
	return &MixinErrChan{errCh: errCh}
}

func (o MixinErrChan) apply(r *Rutina) {
	r.errCh = o.errCh
}
