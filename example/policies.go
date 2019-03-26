// +build ignore

package main

import (
	"context"
	"errors"
	"github.com/neonxp/rutina"
	"log"
	"time"
)

func main() {
	// New instance with builtin context
	r := rutina.New()

	r.Go(func(ctx context.Context) error {
		<-time.After(1 * time.Second)
		log.Println("Do something 1 second without errors and restart")
		return nil
	}, rutina.RestartIfDone, rutina.ShutdownIfFail)

	r.Go(func(ctx context.Context) error {
		<-time.After(2 * time.Second)
		log.Println("Do something 2 seconds without errors and do nothing")
		return nil
	}, rutina.ShutdownIfFail)

	r.Go(func(ctx context.Context) error {
		<-time.After(3 * time.Second)
		log.Println("Do something 3 seconds with error and restart")
		return errors.New("Error!")
	}, rutina.RestartIfFail)

	r.Go(func(ctx context.Context) error {
		<-time.After(4 * time.Second)
		log.Println("Do something 4 seconds with error and do nothing")
		return errors.New("Error!")
	}, rutina.DoNothingIfFail)

	r.Go(func(ctx context.Context) error {
		<-time.After(10 * time.Second)
		log.Println("Do something 10 seconds with error and close context")
		return errors.New("Successfully shutdown at proper place")
	}, rutina.ShutdownIfFail)

	// OS signals subscriber
	r.ListenOsSignals()

	if err := r.Wait(); err != nil {
		log.Fatal(err)
	} else {
		log.Println("Routines stopped but not correct")
	}
}
