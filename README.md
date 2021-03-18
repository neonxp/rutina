# rutina

[![Go Reference](https://pkg.go.dev/badge/github.com/neonxp/rutina.svg)](https://pkg.go.dev/github.com/neonxp/rutina)

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
r := rutina.New()
```

or with custom options:

```go
r := rutina.New(
        ParentContext(ctx context.Context),            // Pass parent context to Rutina (otherwise it uses own new context)
        ListenOsSignals(listenOsSignals ...os.Signal), // Auto listen OS signals and close context on Kill, Term signal
        Logger(l logger),                              // Pass logger for debug, i.e. `log.Printf`
        Errors(errCh chan error),                      // Set errors channel for errors from routines in Restart/DoNothing errors policy
)
```

### Start new routine

```go
r.Go(func (ctx context.Context) error {
    ...do something...
})
```

#### Run Options

```go
r.Go(
    func (ctx context.Context) error {
        ...do something...
    },
    SetOnDone(policy Policy),           // Run policy if returns no error (default: Shutdown)
    SetOnError(policy Policy),          // Run policy if returns error (default: Shutdown)
    SetTimeout(timeout time.Duration),  // Timeout to routine (after it context will be closed)
    SetMaxCount(maxCount int),          // Max tries on Restart policy
)
```

#### Run policies

* `DoNothing` - do not affect other routines
* `Restart` - restart current routine
* `Shutdown` - shutdown all routines

#### Example of run policies

```go
r.Go(func(ctx context.Context) error {
	// If this routine produce no error - all other routines will shutdown (because context cancels)
	// If it returns error - all other routines will shutdown (because context cancels)
},)

r.Go(func(ctx context.Context) error {
	// If this routine produce no error - it restarts
	// If it returns error - all other routines will shutdown (because context cancels)
}, SetOnDone(rutina.Restart))

r.Go(func(ctx context.Context) error {
	// If this routine produce no error - all other routines will shutdown (because context cancels)
	// If it returns error - it will be restarted (maximum 10 times)
}, SetOnError(rutina.Restart), SetMaxCount(10))

r.Go(func(ctx context.Context) error {
	// If this routine stopped by any case other routines will work as before.
}, SetOnDone(rutina.DoNothing))

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
