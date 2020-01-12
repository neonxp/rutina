# rutina

[![GoDoc](https://godoc.org/github.com/neonxp/rutina?status.svg)](https://godoc.org/github.com/neonxp/rutina)

Package Rutina (russian "рутина" - ordinary boring everyday work) is routine orchestrator for your application.

It seems like https://godoc.org/golang.org/x/sync/errgroup with some different:

1) propagates context to every routines. So routine can check if context stopped (`ctx.Done()`).
2) has flexible run/stop policy. i.e. one routine restarts when it errors (useful on daemons) but if errors another - all routines will be cancelled 
3) already has optional signal handler `ListenOsSignals()`

## When it need?

Usually, when your program consists of several routines (i.e.: http server, metrics server and os signals subscriber) and you want to stop all routines when one of them ends (i.e.: by TERM os signal in signal subscriber).

## Usage

### New instance

With default options:

```go
r := rutina.New(nil)
```

or with custom options:

```go
r := rutina.New(
    rutina.Opt.
        SetParentContext(ctx context.Context).      // Pass parent context to Rutina (otherwise it uses own new context)
        SetListenOsSignals(listenOsSignals bool).   // Auto listen OS signals and close context on Kill, Term signal
        SetLogger(l logger).                        // Pass logger for debug, i.e. `log.Printf`
        SetErrors(errCh chan error)                 // Set errors channel for errors from routines in Restart/DoNothing errors policy
)
```

### Start new routine

```go
r.Go(func (ctx context.Context) error {
    ...do something...
}, *runOptions)
```

#### Run Options

```go
RunOpt.
    SetOnDone(policy Policy).           // Run policy if returns no error
    SetOnError(policy Policy).          // Run policy if returns error
    SetTimeout(timeout time.Duration).  // Timeout to routine (after it context will be closed)
    SetMaxCount(maxCount int)           // Max tries on Restart policy
```

#### Run policies

* `DoNothing` - do not affect other routines
* `Restart` - restart current routine
* `Shutdown` - shutdown all routines

#### Example of run policies

```go
r.Go(func(ctx context.Context) error {
	// If this routine produce no error - it just completes, other routines not affected
	// If it returns error - all other routines will shutdown (because context cancels)
}, nil)

r.Go(func(ctx context.Context) error {
	// If this routine produce no error - it restarts
	// If it returns error - all other routines will shutdown (because context cancels)
}, rutina.RunOpt.SetOnDone(rutina.Restart))

r.Go(func(ctx context.Context) error {
	// If this routine produce no error - all other routines will shutdown (because context cancels)
	// If it returns error - it will be restarted
}, rutina.RunOpt.SetOnDone(rutina.Shutdown).SetOnError(rutina.Restart))

r.Go(func(ctx context.Context) error {
	// If this routine stopped by any case - all other routines will shutdown (because context cancels)
}, rutina.RunOpt.SetOnDone(rutina.Shutdown))

r.ListenOsSignals() // Shutdown all routines by OS signal
```

### Wait routines to complete

```go
err := r.Wait()
```

Here err = error that shutdowns all routines (may be will be changed at future)

### Kill routines

```go
id := r.Go(func (ctx context.Context) error { ... })
...
r.Kill(id) // Closes individual context for #id routine that must shutdown it
```

### List of routines

```go
list := r.Processes() 
```

Returns ids of working routines

### Get errors channel

```go
err := <- r.Errors()
```

Disabled by default. Used when passed errors channel to rutina options

## Example

HTTP server with graceful shutdown [`example/http_server.go`](https://github.com/NeonXP/rutina/blob/master/example/http_server.go)

Different run policies [`example/policies.go`](https://github.com/NeonXP/rutina/blob/master/example/policies.go)
