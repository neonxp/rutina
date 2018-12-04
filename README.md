# rutina

Package Rutina (russian "рутина" - ordinary boring everyday work) works like https://godoc.org/golang.org/x/sync/errgroup with small differences:

1) propagates context to routines
2) cancels context when any routine ends with any result (not only when error result)

## When it need?

Usually, when yout program consists of several routines (i.e.: http server, metrics server and os signals subscriber) and you want to stop all routines when one of them ends (i.e.: by TERM os signal in signal subscriber).

## Example

HTTP server with graceful shutdown (`example/http_server.go`):

```
// New instance with builtin context. Alternative: r, ctx := rutina.WithContext(ctx)
r := rutina.New()

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

// OS signals subscriber
r.Go(func(ctx context.Context) error {
    sig := make(chan os.Signal, 1)
    signal.Notify(sig, syscall.SIGTERM, syscall.SIGINT)
    select {
    case <-sig:
        log.Println("TERM or INT signal received")
    case <-ctx.Done():
    }
    return nil
})

if err := r.Wait(); err != nil {
    log.Fatal(err)
}

log.Println("All routines successfully stopped")
```