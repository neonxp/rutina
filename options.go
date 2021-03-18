package rutina

import (
	"context"
	"os"
	"time"
)

type Options struct {
	ParentContext   context.Context
	ListenOsSignals []os.Signal
	Logger          func(format string, v ...interface{})
	Errors          chan error
}

func ParentContext(ctx context.Context) Options {
	return Options{
		ParentContext: ctx,
	}
}

func ListenOsSignals(signals ...os.Signal) Options {
	return Options{
		ListenOsSignals: signals,
	}
}

func Logger(l logger) Options {
	return Options{
		Logger: l,
	}
}

func Errors(errCh chan error) Options {
	return Options{
		Errors: errCh,
	}
}

func composeOptions(opts []Options) Options {
	res := Options{
		ParentContext:   context.Background(),
		Logger:          nopLogger,
		ListenOsSignals: []os.Signal{},
	}
	for _, o := range opts {
		if o.ParentContext != nil {
			res.ParentContext = o.ParentContext
		}
		if o.Errors != nil {
			res.Errors = o.Errors
		}
		if o.ListenOsSignals != nil {
			res.ListenOsSignals = o.ListenOsSignals
		}
		if o.Logger != nil {
			res.Logger = o.Logger
		}
	}
	return res
}

type Policy int

const (
	DoNothing Policy = iota
	Shutdown
	Restart
)

type RunOptions struct {
	OnDone   Policy
	OnError  Policy
	Timeout  *time.Duration
	MaxCount *int
}

func OnDone(policy Policy) RunOptions {
	return RunOptions{
		OnDone: policy,
	}
}

func OnError(policy Policy) RunOptions {
	return RunOptions{
		OnError: policy,
	}
}

func Timeout(timeout time.Duration) RunOptions {
	return RunOptions{
		Timeout: &timeout,
	}
}

func MaxCount(maxCount int) RunOptions {
	return RunOptions{
		MaxCount: &maxCount,
	}
}

func composeRunOptions(opts []RunOptions) RunOptions {
	res := RunOptions{
		OnDone:  Shutdown,
		OnError: Shutdown,
	}
	for _, o := range opts {
		if o.OnDone != res.OnDone {
			res.OnDone = o.OnDone
		}
		if o.OnError != res.OnError {
			res.OnError = o.OnError
		}
		if o.MaxCount != nil {
			res.MaxCount = o.MaxCount
		}
		if o.Timeout != nil {
			res.Timeout = o.Timeout
		}
	}
	return res
}
