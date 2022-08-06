package rutina

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestSuccess(t *testing.T) {
	r := New()
	counter := 0
	f := func(name string, ttl time.Duration) error {
		counter++
		<-time.After(ttl)
		counter--
		t.Log(name)
		return nil
	}
	r.Go(func(ctx context.Context) error {
		return f("one", 1*time.Second)
	})
	r.Go(func(ctx context.Context) error {
		return f("two", 2*time.Second)
	})
	r.Go(func(ctx context.Context) error {
		return f("three", 3*time.Second)
	})
	if err := r.Wait(); err != nil {
		t.Error("Unexpected error", err)
	}
	if counter == 0 {
		t.Log("All routines done")
	} else {
		t.Error("Not all routines stopped")
	}
}

func TestError(t *testing.T) {
	r := New()
	f := func(name string, ttl time.Duration) error {
		<-time.After(ttl)
		t.Log(name)
		return errors.New("error from " + name)
	}
	r.Go(func(ctx context.Context) error {
		return f("one", 1*time.Second)
	})
	r.Go(func(ctx context.Context) error {
		return f("two", 2*time.Second)
	})
	r.Go(func(ctx context.Context) error {
		return f("three", 3*time.Second)
	})
	if err := r.Wait(); err != nil {
		if err.Error() != "error from one" {
			t.Error("Must be error from first routine")
		}
		t.Log(err)
	}
	t.Log("All routines done")
}

func TestErrorWithRestart(t *testing.T) {
	maxCount := 2

	r := New()
	r.Go(func(ctx context.Context) error {
		return nil
	})
	r.Go(func(ctx context.Context) error {
		return errors.New("error")
	}, RunOptions{
		OnError:  Restart,
		MaxCount: &maxCount,
	})

	err := r.Wait()
	if err != ErrRunLimit {
		t.Error("Must be an error ErrRunLimit from r.Wait since all restarts was executed")
	}
}

func TestContext(t *testing.T) {
	r := New()
	cc := false
	r.Go(func(ctx context.Context) error {
		<-time.After(1 * time.Second)
		return nil
	}, RunOptions{OnDone: Shutdown})
	r.Go(func(ctx context.Context) error {
		select {
		case <-ctx.Done():
			cc = true
			return nil
		case <-time.After(3 * time.Second):
			return errors.New("Timeout")
		}
	})
	if err := r.Wait(); err != nil {
		t.Error("Unexpected error", err)
	}
	if cc {
		t.Log("Second routine successfully complete by context done")
	} else {
		t.Error("Routine not completed by context")
	}
}
