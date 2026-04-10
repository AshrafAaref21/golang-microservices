package main

import (
	"context"
	"log"
	"net"
	"os"
	"os/signal"
	"ride-sharing/services/trip-service/internal/infrastructure/events"
	"ride-sharing/services/trip-service/internal/infrastructure/grpc"
	"ride-sharing/services/trip-service/internal/infrastructure/repository"
	"ride-sharing/services/trip-service/internal/service"
	"ride-sharing/shared/env"
	"ride-sharing/shared/messaging"
	"syscall"

	grpcserver "google.golang.org/grpc"
)

var GrpcAddr = ":9083"

func main() {
	rabbitMqURI := env.GetString("RABBITMQ_URI", "amqp://guest:guest@rabbitmq:5672/")

	inmemoryRepo := repository.NewinMemoryRepository()
	service := service.NewTripService(inmemoryRepo)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		signalChan := make(chan os.Signal, 1)
		signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)
		<-signalChan
		log.Println("Received interrupt signal, shutting down...")
		cancel()
	}()

	lis, err := net.Listen("tcp", GrpcAddr)
	if err != nil {
		log.Fatalf("Failed to listen on %s: %v", GrpcAddr, err)
	}

	rabbitmq, err := messaging.NewRabbitMQ(rabbitMqURI)
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %v", err)
	}
	defer rabbitmq.Close()
	err = rabbitmq.SetupExchangesAndQueues()
	if err != nil {
		log.Fatalf("Failed to set up RabbitMQ exchanges and queues: %v", err)
	}
	log.Printf("Connected to RabbitMQ at %s", rabbitMqURI)

	eventPublisher := events.NewTripEventPublisher(rabbitmq)

	grpcServer := grpcserver.NewServer()
	grpc.NewGrpcHandler(grpcServer, service, eventPublisher)

	log.Printf("Trip Service gRPC server is running on %s", GrpcAddr)

	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("Failed to serve gRPC server: %v", err)
			cancel()
		}
	}()

	<-ctx.Done()
	log.Println("Shutting down gRPC server...")
	grpcServer.GracefulStop()
}
