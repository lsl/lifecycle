package main

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/lsl/lifecycle"
)

func main() {
	l := lifecycle.New()
	defer l.Shutdown()

	l.TrapSignals()

	initServer(l)
	initWorker(l)

	<-l.Done()
	fmt.Println("gracefully shut down")
}

func initServer(l *lifecycle.Lifecycle) {
	srv := &http.Server{Addr: ":8080"}

	go func() {
		fmt.Println("HTTP server listening on :8080")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Println("server error:", err)
		}
	}()

	l.OnShutdown(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		fmt.Println("shutting down HTTP server")
		srv.Shutdown(ctx)
	})
}

func initWorker(l *lifecycle.Lifecycle) {
	done := make(chan struct{})

	go func() {
		ticker := time.NewTicker(500 * time.Millisecond)
		defer ticker.Stop()
		for {
			select {
			case <-l.Done():
				fmt.Println("worker received shutdown signal")
				close(done)
				return
			case <-ticker.C:
				fmt.Println("worker tick")
			}
		}
	}()

	l.OnShutdown(func() {
		<-done
		fmt.Println("worker cleaned up")
	})
}
