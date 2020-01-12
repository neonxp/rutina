// +build ignore

package main

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/neonxp/rutina"
)

func main() {
	// New instance with builtin context
	r := rutina.New(rutina.Opt.SetLogger(log.Printf).SetListenOsSignals(true))

	r.Go(func(ctx context.Context) error {
		<-time.After(1 * time.Second)
		log.Println("Do something 1 second without errors and restart")
		return nil
	}, nil)

	r.Go(func(ctx context.Context) error {
		<-time.After(2 * time.Second)
		log.Println("Do something 2 seconds without errors and do nothing")
		return nil
	}, nil)

	r.Go(func(ctx context.Context) error {
		select {
		case <-time.After(time.Second):
			return errors.New("max 10 times")
		case <-ctx.Done():
			return nil
		}
	}, rutina.RunOpt.SetOnError(rutina.Restart).SetMaxCount(10))

	r.Go(func(ctx context.Context) error {
		select {
		case <-time.After(time.Second):
			return errors.New("max 10 seconds")
		case <-ctx.Done():
			return nil
		}
	}, rutina.RunOpt.SetOnError(rutina.Restart).SetTimeout(10*time.Second))

	if err := r.Wait(); err != nil {
		log.Fatal(err)
	} else {
		log.Println("Routines stopped")
	}
}
