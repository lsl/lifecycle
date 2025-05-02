package lifecycle

import (
	"context"
	"sync"
	"testing"
	"time"
)

func TestLifecycle_ShutdownRunsHooksInReverseOrder(t *testing.T) {
	l := New()
	var mu sync.Mutex
	var order []int

	// Register two shutdown hooks that record their order of execution
	l.OnShutdown(func() {
		mu.Lock()
		defer mu.Unlock()
		order = append(order, 1)
	})
	l.OnShutdown(func() {
		mu.Lock()
		defer mu.Unlock()
		order = append(order, 2)
	})

	l.Shutdown()

	// Hooks should run in reverse order: [2, 1]
	if len(order) != 2 || order[0] != 2 || order[1] != 1 {
		t.Fatalf("expected shutdown order [2 1], got %v", order)
	}
}

func TestLifecycle_ShutdownOnlyOnce(t *testing.T) {
	l := New()
	var count int

	// Register a hook and call Shutdown() twice
	l.OnShutdown(func() { count++ })
	l.Shutdown()
	l.Shutdown()

	// Hook should only run once
	if count != 1 {
		t.Fatalf("expected hook to run once, ran %d times", count)
	}
}

func TestLifecycle_ContextCancelledOnShutdown(t *testing.T) {
	l := New()
	done := make(chan struct{})

	// Wait for the Done channel to be closed
	go func() {
		<-l.Done()
		close(done)
	}()

	l.Shutdown()

	// Done should be closed promptly after shutdown
	select {
	case <-done:
		// ok
	case <-time.After(100 * time.Millisecond):
		t.Fatal("context was not cancelled on shutdown")
	}
}

func TestLifecycle_ShutdownOnContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	l := WithContext(ctx)
	done := make(chan struct{})

	// Register a shutdown hook that signals completion
	l.OnShutdown(func() {
		close(done)
	})

	// Link shutdown to the given context
	l.ShutdownOn(ctx)
	cancel()

	// Shutdown should be triggered by ctx cancellation
	select {
	case <-done:
		// ok
	case <-time.After(100 * time.Millisecond):
		t.Fatal("ShutdownOn did not trigger shutdown")
	}
}

func TestLifecycle_PanicIfRegisterAfterShutdown(t *testing.T) {
	l := New()
	l.Shutdown()

	// Expect panic if we register after shutdown
	defer func() {
		recoverVal := recover()
		if recoverVal == nil {
			t.Fatal("expected panic when registering after shutdown")
		}
	}()

	l.OnShutdown(func() {})
}
