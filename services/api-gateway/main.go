package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"ride-sharing/shared/env"
	"ride-sharing/shared/messaging"
)

var (
	httpAddr    = env.GetString("HTTP_ADDR", ":8081")
	RabbitMQURI = env.GetString("RABBITMQ_URI", "amqp://guest:guest@rabbitmq:5672/")
)

func main() {
	log.Println("Starting API Gateway")
	mux := http.NewServeMux()
	rabbitmq, err := messaging.NewRabbitMQ(RabbitMQURI)
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %v", err)
	}
	defer rabbitmq.Close()
	if err := rabbitmq.SetupExchangesAndQueues(); err != nil {
		log.Fatalf("Failed to setup RabbitMQ: %v", err)
	}

	mux.HandleFunc("POST /trip/preview", handleTripPreview)
	mux.HandleFunc("POST /trip/start", handleTripStart)
	mux.HandleFunc("/ws/drivers", func(w http.ResponseWriter, r *http.Request) {
		handleDriversWebSocket(w, r, rabbitmq)
	})
	mux.HandleFunc("/ws/riders", func(w http.ResponseWriter, r *http.Request) {
		handleRidersWebSocket(w, r, rabbitmq)
	})
	mux.HandleFunc("/webhook/stripe", func(w http.ResponseWriter, r *http.Request) {
		handleStripeWebhook(w, r, rabbitmq)
	})

	server := &http.Server{
		Addr:    httpAddr,
		Handler: enableCORS(mux),
	}

	serverErrors := make(chan error, 1)
	go func() {
		log.Printf("API Gateway is running on %s", httpAddr)
		serverErrors <- server.ListenAndServe()
	}()

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt)

	select {
	case err := <-serverErrors:
		log.Fatalf("Could not start server: %v", err)
	case sig := <-shutdown:
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		log.Printf("Received signal %v, shutting down gracefully...", sig)
		if err := server.Shutdown(ctx); err != nil {
			log.Printf("Graceful shutdown failed: %v", err)
			server.Close()
		}
	}
}
