// +build ignore

package main

import (
	"context"
	"errors"
	"log"
	"os"
	"time"

	"github.com/neonxp/rutina/v3"
)

func main() {
	// New instance with builtin context
	r := rutina.New(rutina.Logger(log.Printf), rutina.ListenOsSignals(os.Interrupt, os.Kill))

	r.Go(func(ctx context.Context) error {
		<-time.After(1 * time.Second)
		log.Println("Do something 1 second without errors and restart")
		return nil
	})

	r.Go(func(ctx context.Context) error {
		<-time.After(2 * time.Second)
		log.Println("Do something 2 seconds without errors and do nothing")
		return nil
	})

	r.Go(func(ctx context.Context) error {
		select {
		case <-time.After(time.Second):
			return errors.New("max 10 times")
		case <-ctx.Done():
			return nil
		}
	}, rutina.OnError(rutina.Restart), rutina.MaxCount(10))

	r.Go(func(ctx context.Context) error {
		select {
		case <-time.After(time.Second):
			return errors.New("max 10 seconds")
		case <-ctx.Done():
			return nil
		}
	}, rutina.OnError(rutina.Restart), rutina.SetTimeout(10*time.Second))

	if err := r.Wait(); err != nil {
		log.Fatal(err)
	} else {
		log.Println("Routines stopped")
	}
}
