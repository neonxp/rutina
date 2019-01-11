// +build ignore

package main

import (
	"context"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/neonxp/rutina"
)

func main() {
	// New instance with builtin context. Alternative: r, ctx := rutina.WithContext(ctx)
	r, _ := rutina.New()

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
}
