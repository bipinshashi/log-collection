package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	handler "github.com/bipinshashi/log-collection/internal"
	"github.com/bipinshashi/log-collection/internal/config"
	"github.com/gorilla/mux"
)

func main() {

	client := &http.Client{
		Timeout: time.Second * 10,
	}

	appHandler := &handler.AppHandler{
		Client: client,
	}

	r := mux.NewRouter()
	r.HandleFunc("/api/v1/logs", appHandler.GetLogs).Methods("GET")
	r.HandleFunc("/", appHandler.ShowDemo).Methods("GET")

	config := config.GetConfig()

	srv := &http.Server{
		Addr: "0.0.0.0:" + config.Port,
		// Good practice to set timeouts to avoid Slowloris attacks.
		WriteTimeout: time.Second * 15,
		ReadTimeout:  time.Second * 15,
		IdleTimeout:  time.Second * 60,
		Handler:      r,
	}

	// Run our server in a goroutine so that it doesn't block.
	go func() {
		if err := srv.ListenAndServe(); err != nil {
			log.Println(err)
		}
	}()
	log.Printf("Server started on port %s", config.Port)

	var wait time.Duration

	c := make(chan os.Signal, 1)
	// We'll accept graceful shutdowns when quit via SIGINT (Ctrl+C)
	// SIGKILL, SIGQUIT or SIGTERM (Ctrl+/) will not be caught.
	signal.Notify(c, os.Interrupt)

	// Block until we receive our signal.
	<-c

	// Create a deadline to wait for.
	ctx, cancel := context.WithTimeout(context.Background(), wait)
	defer cancel()
	// Doesn't block if no connections, but will otherwise wait
	// until the timeout deadline.
	srv.Shutdown(ctx)
	log.Println("shutting down")
	os.Exit(0)
}
