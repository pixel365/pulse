package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/joho/godotenv"

	"github.com/pixel365/pulse/internal/api"

	"github.com/pixel365/pulse/internal/db/postgres"
	"github.com/pixel365/pulse/internal/logger"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	_ = godotenv.Load()

	log := logger.NewSlog()

	pg := postgres.NewConfigFromEnv()
	pool, err := postgres.NewPool(ctx, pg)
	if err != nil {
		log.Error(ctx, "postgres pool error", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(api.RequestLogger(log))

	r.Route("/v1", func(r chi.Router) {
		r.Get("/services", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotImplemented)
		})

		r.Get("/services/{serviceId}/checks/state", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotImplemented)
		})

		r.Get(
			"/services/{serviceId}/checks/{checkId}/executions",
			func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusNotImplemented)
			},
		)

		r.Get(
			"/services/{serviceId}/checks/{checkId}/timeline",
			func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusNotImplemented)
			},
		)
	})

	addr := os.Getenv("API_LISTEN_ADDR")
	if addr == "" {
		log.Warn(ctx, "API_LISTEN_ADDR is not set. Default is :8080")
		addr = ":8080"
	}

	server := &http.Server{
		Addr:              addr,
		ReadHeaderTimeout: 2 * time.Second,
		ReadTimeout:       5 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       30 * time.Second,
		Handler:           r,
	}

	go func() {
		if listenErr := server.ListenAndServe(); listenErr != nil &&
			!errors.Is(listenErr, http.ErrServerClosed) {
			log.Error(ctx, "http server error", "error", listenErr)
			os.Exit(1)
		}
	}()

	<-ctx.Done()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err = server.Shutdown(shutdownCtx); err != nil {
		log.Error(ctx, "http server shutdown error", "error", err)
		os.Exit(1)
	}
}
