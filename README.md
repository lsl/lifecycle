# lifecycle

A tiny Go package that helps manage graceful application shutdowns using context and ordered cleanup hooks.

## 🧠 Why use `lifecycle`?

Go's built-in `defer` works well for simple teardown. But once you start managing:

- concurrent goroutines
- servers or background workers
- complex test setups
- signal handling
- shutdown from a worker

...you need teardown control that isn't tied to function scope.

`lifecycle` gives you a single place to:
- track your app's cancellation context
- register teardown hooks in reverse-order
- hook into OS signals or other contexts

## 🚀 Usage

```go
import "github.com/lsl/lifecycle"

func main() {
	l := lifecycle.New()
	defer l.Shutdown()

	l.TrapSignals() // optional: handle SIGINT/SIGTERM

	l.OnShutdown(func() {
		fmt.Println("cleaning up db")
	})

	l.OnShutdown(func() {
		fmt.Println("stopping server")
	})

	<-l.Done()
	fmt.Println("bye")
}
```

## 🔌 Integrating with subsystems

```go
func InitServer(ctx context.Context, l *lifecycle.Lifecycle) {
	srv := &http.Server{Addr: ":8080"}

	go srv.ListenAndServe()

	l.OnShutdown(func() {
		srv.Shutdown(context.Background())
	})
}
```

## 🧪 Testing

You can manually trigger shutdown via `l.Shutdown()` or tie it to a timeout:

```go
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

l := lifecycle.WithContext(ctx)
l.ShutdownOn(ctx)
```

## 🧼 Hook ordering

Hooks are run in **reverse order** of registration, just like `defer`. This allows dependency-aware teardown:

```go
l.OnShutdown(closeDB)
l.OnShutdown(stopServer)
// stopServer runs before closeDB
```

## 📦 API

- `lifecycle.New()`
- `lifecycle.WithContext(ctx)`
- `l.OnShutdown(func())` — register a cleanup function
- `l.Shutdown()` — cancels context and runs hooks
- `l.Done()` — the `context.Done()` channel
- `l.TrapSignals()` — listen for SIGINT/SIGTERM
- `l.ShutdownOn(ctx)` — shutdown when a context ends

