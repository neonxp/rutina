package rutina

import (
	"context"
	"time"
)

type Options struct {
	ParentContext   context.Context
	ListenOsSignals bool
	Logger          func(format string, v ...interface{})
	Errors          chan error
}

func (o *Options) SetParentContext(ctx context.Context) *Options {
	o.ParentContext = ctx
	return o
}

func (o *Options) SetListenOsSignals(listenOsSignals bool) *Options {
	o.ListenOsSignals = listenOsSignals
	return o
}

func (o *Options) SetLogger(l logger) *Options {
	o.Logger = l
	return o
}

func (o *Options) SetErrors(errCh chan error) *Options {
	o.Errors = errCh
	return o
}

var Opt = &Options{
	ParentContext:   context.Background(),
	ListenOsSignals: false,
	Logger:          nil,
	Errors:          nil,
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

func (rp *RunOptions) SetOnDone(policy Policy) *RunOptions {
	rp.OnDone = policy
	return rp
}

func (rp *RunOptions) SetOnError(policy Policy) *RunOptions {
	rp.OnError = policy
	return rp
}

func (rp *RunOptions) SetTimeout(timeout time.Duration) *RunOptions {
	rp.Timeout = &timeout
	return rp
}

func (rp *RunOptions) SetMaxCount(maxCount int) *RunOptions {
	rp.MaxCount = &maxCount
	return rp
}

var RunOpt = &RunOptions{
	OnDone:   DoNothing,
	OnError:  Shutdown,
	Timeout:  nil,
	MaxCount: nil,
}
