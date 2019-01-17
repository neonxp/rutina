package rutina

import (
	"context"
	"log"
	"os"
)

type Option interface {
	apply(*Rutina)
}

type OptionContext struct {
	Context context.Context
}

func WithContext(context context.Context) *OptionContext {
	return &OptionContext{Context: context}
}

func (o OptionContext) apply(r *Rutina) {
	ctx, cancel := context.WithCancel(o.Context)
	r.ctx = ctx
	r.Cancel = cancel
}

type OptionLogger struct {
	Logger *log.Logger
}

func WithLogger(logger *log.Logger) *OptionLogger {
	return &OptionLogger{Logger: logger}
}

func WithStdLogger() *OptionLogger {
	return &OptionLogger{Logger: log.New(os.Stdout, "rutina", log.LstdFlags)}
}

func (o OptionLogger) apply(r *Rutina) {
	r.logger = o.Logger
}

type OptionCancelByError struct{}

func WithCancelByError() *OptionCancelByError {
	return &OptionCancelByError{}
}

func (OptionCancelByError) apply(r *Rutina) {
	r.cancelByError = true
}
