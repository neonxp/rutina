// +build ignore

package main

import (
	"context"
	"github.com/neonxp/rutina"
	"io"
	"log"
	"net/http"
)

func main() {
	// New instance with builtin context. Alternative: r, ctx := rutina.OptionContext(ctx)
	r, _ := rutina.New(rutina.WithStdLogger())

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
	r.ListenOsSignals()

	if err := r.Wait(); err != nil {
		log.Fatal(err)
	}
	log.Println("All routines successfully stopped")
}
