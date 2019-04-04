# rutina

[![GoDoc](https://godoc.org/github.com/neonxp/rutina?status.svg)](https://godoc.org/github.com/neonxp/rutina)

Package Rutina (russian "рутина" - ordinary boring everyday work) is routine orchestrator for your application.

It seems like https://godoc.org/golang.org/x/sync/errgroup with some different:

1) propagates context to every routines. So routine can check if context stopped (`ctx.Done()`).
2) has flexible run/stop policy. i.e. one routine restarts when it fails (useful on daemons) but if fails another - all routines will be cancelled 
3) already has optional signal handler `ListenOsSignals()`

## When it need?

Usually, when your program consists of several routines (i.e.: http server, metrics server and os signals subscriber) and you want to stop all routines when one of them ends (i.e.: by TERM os signal in signal subscriber).

## Usage

### New instance

```go
r := rutina.New()
```

or with optional mixins (see below):

```go
r := rutina.New(...Mixins)
```
or
```go 
r.With(...Mixins)
```

### Start new routine

```go
r.Go(func (ctx context.Context) error {
    ...do something...
}, ...runOptions)
```

Available options of run policy:

* `ShutdownIfFail` - Shutdown all routines if this routine fails
* `RestartIfFail` - Restart this routine if it fail
* `DoNothingIfFail` - Do nothing just stop this routine if it fail
* `ShutdownIfDone` - Shutdown all routines if this routine done without errors
* `RestartIfDone` - Restart if this routine done without errors
* `DoNothingIfDone` - Do nothing if this routine done without errors

Default policy:

`ShutdownIfFail` && `ShutdownIfDone`

(just like [errgroup](https://godoc.org/golang.org/x/sync/errgroup)) 

#### Example of run policies

```go
r.Go(func(ctx context.Context) error {
	// If this routine produce no error - it just restarts
	// If it returns error - all other routines will shutdown (because context cancels)
}, rutina.RestartIfDone, rutina.ShutdownIfFail)

r.Go(func(ctx context.Context) error {
	// If this routine produce no error - it just completes
	// If it returns error - all other routines will shutdown (because context cancels)
}, rutina.DoNothingIfDone, rutina.ShutdownIfFail)

r.Go(func(ctx context.Context) error {
	// If this routine produce no error - all other routines will shutdown (because context cancels)
	// If it returns error - it will be restarted
}, rutina.RestartIfFail)

r.Go(func(ctx context.Context) error {
	// If this routine stopped by any case - all other routines will shutdown (because context cancels)
}, rutina.ShutdownIfDone)

r.ListenOsSignals() // Shutdown all routines by OS signal
```

### Wait routines to complete

```go
err := r.Wait()
```

Here err = error that shutdowns all routines (may be will be changed at future)

### Get errors channel

```go
err := <- r.Errors()
```

Disabled by default. Use `r.With(rutina.WithErrChan())` to turn on.

## Mixins

### Usage

```go
r := rutina.New(mixin1, mixin2, ...)
```
or
```go
r := rutina.New()
r = r.With(mixin1, mixin2, ...) // Returns new instance of Rutina!
```

### Logger

```go 
r = r.With(rutina.WithStdLogger())
``` 
or 
```go 
r = r.With(rutina.WithLogger(logger log.Logger))
```

Sets standard or custom logger. By default there is no logger.

### Custom context

```go
r = r.With(rutina.WithContext(ctx context.Context))
````

Propagates your own context to Rutina. By default it use own context. 

### Enable errors channel

```go
r = r.With(rutina.WithErrChan())
...
err := <- r.Errors()
```

Turn on errors channel

## Example

HTTP server with graceful shutdown [`example/http_server.go`](https://github.com/NeonXP/rutina/blob/master/example/http_server.go)

Different run policies [`example/policies.go`](https://github.com/NeonXP/rutina/blob/master/example/policies.go)
