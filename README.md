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

* `ShutdownIfError` - Shutdown all routines if this routine returns error
* `RestartIfError` - Restart this routine if this routine returns error
* `DoNothingIfError` - Do nothing just stop this routine if this routine returns error
* `ShutdownIfDone` - Shutdown all routines if this routine done without errors
* `RestartIfDone` - Restart if this routine done without errors
* `DoNothingIfDone` - Do nothing if this routine done without errors

Default policy:

`ShutdownIfError` && `ShutdownIfDone`

(just like [errgroup](https://godoc.org/golang.org/x/sync/errgroup)) 

#### Example of run policies

```go
r.Go(func(ctx context.Context) error {
	// If this routine produce no error - it just restarts
	// If it returns error - all other routines will shutdown (because context cancels)
}, rutina.RestartIfDone, rutina.ShutdownIfError)

r.Go(func(ctx context.Context) error {
	// If this routine produce no error - it just completes
	// If it returns error - all other routines will shutdown (because context cancels)
}, rutina.DoNothingIfDone, rutina.ShutdownIfError)

r.Go(func(ctx context.Context) error {
	// If this routine produce no error - all other routines will shutdown (because context cancels)
	// If it returns error - it will be restarted
}, rutina.RestartIfError)

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

## Lifecycle events

Rutina has own simple lifecycle events:

* `EventRoutineStart` - Fires when starts new routine
* `EventRoutineStop` - Fires when routine stopped with any result 
* `EventRoutineComplete` - Fires when routine stopped without errors
* `EventRoutineError` - Fires when routine stopped with error
* `EventAppStop` - Fires when all routines stopped with any result
* `EventAppComplete` - Fires when all routines stopped with no errors
* `EventAppError` - Fires when all routines stopped with error

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

### Lifecycle listener

```go
r = r.With(rutina.WithLifecycleListener(func (event rutina.Event, rid int) { ... }))
```

Simple lifecycle listener

### Auto listen OS signals

```go
r = r.With(rutina.WithListenOsSignals())
```

Automatically listen OS signals. There is no `r.ListenOsSignals()` needed.

## Example

HTTP server with graceful shutdown [`example/http_server.go`](https://github.com/NeonXP/rutina/blob/master/example/http_server.go)

Different run policies [`example/policies.go`](https://github.com/NeonXP/rutina/blob/master/example/policies.go)
