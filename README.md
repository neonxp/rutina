# rutina

Package Rutina (russian "рутина" - ordinary boring everyday work) is routine orchestrator for your application.

It seems like https://godoc.org/golang.org/x/sync/errgroup with some different:

1) propagates context to every routines. So routine can check if context stopped (`ctx.Done()`).
2) by default cancels context when any routine ends with any result (not only when error result). Can be configured by option `OptionCancelByError`.
3) already has optional signal handler `ListenOsSignals()`

## When it need?

Usually, when your program consists of several routines (i.e.: http server, metrics server and os signals subscriber) and you want to stop all routines when one of them ends (i.e.: by TERM os signal in signal subscriber).

## Usage

### New instance

`r := rutina.New()`

or with options (see below):

`r := rutina.New(...Option)` or `r.WithOptions(...Option)`

### Start new routine

```
r.Go(func (ctx context.Context) error {
    ...do something...
})
```

### Wait routines to complete

```
err := r.Wait()
```

Here err = first error in any routine

## Options

### Usage options

`r := rutina.New(option1, option2, ...)`
or
```
r := rutina.New()
r = r.WithOptions(option1, option2, ...) // Returns new instance of Rutina!
```

### Logger

`rutina.WithLogger(logger log.Logger) Option` or `rutina.WithStdLogger() Option`

### Custom context

`rutina.WithContext(ctx context.Context) Option`

### Cancel only by errors

`rutina.WithCancelByError() Option`

If this option set, rutina doesnt cancel context if routine completed without error.

## Example

HTTP server with graceful shutdown (`example/http_server.go`):

```
// New instance with builtin context. Alternative: r, ctx := rutina.WithContext(ctx)
r := rutina.New(rutina.WithStdLogger())

srv := &http.Server{Addr: ":8080"}
http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
    io.WriteString(w, "hello world\n")
})

// Starting http server and listen connections
r.Go(func(ctx context.Context) error {
    if err := srv.ListenAndServe(); err != nil {
        return err
    }
    log.Println("Server stopped")
    return nil
})

// Gracefully stoping server when context canceled
r.Go(func(ctx context.Context) error {
    <-ctx.Done()
    log.Println("Stopping server...")
    return srv.Shutdown(ctx)
})

// OS signals listener
r.ListenOsSignals()

if err := r.Wait(); err != nil {
    log.Fatal(err)
}

log.Println("All routines successfully stopped")
```