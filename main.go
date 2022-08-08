package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/jace-ys/simple-api/httpapi"
	"github.com/jace-ys/simple-api/server"
)

var (
	port = flag.Int("port", 8000, "Port binding for the HTTP server.")
)

func main() {
	ctx, stop := signal.NotifyContext(context.TODO(), os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", *port),
		Handler: handler(),
	}

	go func() {
		<-ctx.Done()
		stop()

		ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
		defer cancel()

		log.Println("attempting graceful shutdown")
		if err := srv.Shutdown(ctx); err != nil {
			log.Fatalf("server shutdown error: %s\n", err)
		}
	}()

	log.Printf("server listening on %s\n", srv.Addr)
	if err := srv.ListenAndServe(); err != http.ErrServerClosed {
		log.Fatalf("server failed to serve: %s\n", err)
	}

	log.Println("server stopped")
}

func handler() http.Handler {
	router := mux.NewRouter()
	router.Handle("/metrics", promhttp.Handler())

	v1 := router.PathPrefix("/api/v1/").Subrouter()

	{
		router := v1.PathPrefix("/mcu").Subrouter()
		handler := server.NewMCUHandler(httpapi.NewMCUClient())
		handler.RegisterRoutes(router)
	}

	return handlers.LoggingHandler(os.Stdout, router)
}
