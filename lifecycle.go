package lifecycle

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

// Lifecycle manages context and registered shutdown hooks.
type Lifecycle struct {
	ctx    context.Context
	cancel context.CancelFunc

	mu     sync.Mutex
	hooks  []func()
	closed bool
}

// New creates a new Lifecycle instance with its own cancellable context.
func New() *Lifecycle {
	ctx, cancel := context.WithCancel(context.Background())
	return &Lifecycle{
		ctx:    ctx,
		cancel: cancel,
	}
}

// WithContext creates a Lifecycle instance bound to an existing context.
func WithContext(ctx context.Context) *Lifecycle {
	child, cancel := context.WithCancel(ctx)
	return &Lifecycle{
		ctx:    child,
		cancel: cancel,
	}
}

// Done returns the Lifecycle instance's context.Done channel.
func (l *Lifecycle) Done() <-chan struct{} {
	return l.ctx.Done()
}

// OnShutdown registers a cleanup hook to be called during Shutdown.
// Hooks are called in reverse order of registration.
func (l *Lifecycle) OnShutdown(f func()) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.closed {
		panic("cannot register: lifecycle already shut down")
	}
	l.hooks = append(l.hooks, f)
}

// Shutdown cancels the context and runs all registered hooks once.
func (l *Lifecycle) Shutdown() {
	l.mu.Lock()
	if l.closed {
		l.mu.Unlock()
		return
	}
	l.closed = true
	hooks := make([]func(), len(l.hooks))
	copy(hooks, l.hooks)
	l.mu.Unlock()

	l.cancel()

	for i := len(hooks) - 1; i >= 0; i-- {
		hooks[i]()
	}
}

// TrapSignals sets up SIGINT and SIGTERM handling to trigger Shutdown.
func (l *Lifecycle) TrapSignals() {
	go func() {
		ch := make(chan os.Signal, 1)
		signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
		<-ch
		l.Shutdown()
	}()
}

// ShutdownOn cancels the Lifecycle instance when the provided context is done.
func (l *Lifecycle) ShutdownOn(ctx context.Context) {
	go func() {
		<-ctx.Done()
		l.Shutdown()
	}()
}

// ExampleLifecycle demonstrates basic usage of Lifecycle.
func ExampleLifecycle() {
	l := New()
	l.TrapSignals()

	l.OnShutdown(func() {
		fmt.Println("cleanup 1")
	})
	l.OnShutdown(func() {
		fmt.Println("cleanup 2")
	})

	go func() {
		time.Sleep(100 * time.Millisecond)
		l.Shutdown()
	}()

	<-l.Done()
	fmt.Println("done")
	// Output:
	// cleanup 2
	// cleanup 1
	// done
}
