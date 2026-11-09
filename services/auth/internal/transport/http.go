package transport

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func StartHTTPServer(ctx context.Context, addr string) error {
	mux := http.NewServeMux()

	mux.Handle("/metrics", promhttp.Handler())
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	})

	srv := &http.Server{
		Addr:              addr,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		<-ctx.Done()

		shutdownCtx, cancel := context.WithTimeout(
			context.Background(),
			5*time.Second,
		)
		defer cancel()

		if err := srv.Shutdown(shutdownCtx); err != nil {
			log.Printf("HTTP server shutdown error: %v", err)
		}
	}()

	log.Printf("HTTP server listening on %s", addr)

	err := srv.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		log.Printf("HTTP server failed: %v", err)
		return err
	}

	return err
}
