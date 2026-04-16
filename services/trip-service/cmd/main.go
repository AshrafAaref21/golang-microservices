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
	"ride-sharing/shared/tracing"
	"syscall"

	grpcserver "google.golang.org/grpc"
)

var GrpcAddr = ":9083"

func main() {
	rabbitMqURI := env.GetString("RABBITMQ_URI", "amqp://guest:guest@rabbitmq:5672/")

	tracerCfg := tracing.Config{
		ServiceName: "trip-service",
		Environment: env.GetString("ENVIRONMENT", "development"),
		Endpoint:    env.GetString("JAEGER_ENDPOINT", "http://jaeger:14268/api/traces"),
	}
	shdw, err := tracing.InitTracer(tracerCfg)
	if err != nil {
		log.Fatalf("Failed to initialize the tracer: %v", err)
	}

	inmemoryRepo := repository.NewinMemoryRepository()
	service := service.NewTripService(inmemoryRepo)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	defer shdw(ctx)

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

	log.Printf("Connected to RabbitMQ at %s", rabbitMqURI)

	eventPublisher := events.NewTripEventPublisher(rabbitmq)
	driverConsumer := events.NewDriverConsumer(rabbitmq, service)

	go func() {
		if err := driverConsumer.Listen(); err != nil {
			log.Printf("Failed to start driver consumer: %v", err)
			cancel()
		}
	}()

	paymentConsumer := events.NewPaymentConsumer(rabbitmq, service)
	go paymentConsumer.Listen()

	grpcServer := grpcserver.NewServer(tracing.WithTracingInterceptors()...)
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
