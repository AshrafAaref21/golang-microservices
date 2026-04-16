package main

import (
	"context"
	"log"
	"net"
	"os"
	"os/signal"
	"ride-sharing/shared/env"
	"ride-sharing/shared/messaging"
	"ride-sharing/shared/tracing"
	"syscall"

	grpcserver "google.golang.org/grpc"
)

var GrpcAddr = ":9092"

func main() {
	tracerCfg := tracing.Config{
		ServiceName: "driver-service",
		Environment: env.GetString("ENVIRONMENT", "development"),
		Endpoint:    env.GetString("JAEGER_ENDPOINT", "http://jaeger:14268/api/traces"),
	}

	shdw, err := tracing.InitTracer(tracerCfg)
	if err != nil {
		log.Fatalf("Failed to initialize the tracer: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	defer shdw(ctx)

	rabbitMqURI := env.GetString("RABBITMQ_URI", "amqp://guest:guest@rabbitmq:5672/")

	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
		<-sigCh
		cancel()
	}()

	lis, err := net.Listen("tcp", GrpcAddr)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	rabbitmq, err := messaging.NewRabbitMQ(rabbitMqURI)
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %v", err)
	}
	log.Printf("Connected to RabbitMQ at %s", rabbitMqURI)
	defer rabbitmq.Close()

	svc := NewService()

	// Starting the gRPC server
	grpcServer := grpcserver.NewServer()
	NewGrpcHandler(grpcServer, svc)

	consumer := NewTripConsumer(rabbitmq, svc)

	go func() {
		if err := consumer.Listen(); err != nil {
			log.Printf("failed to start consumer: %v", err)
			cancel()
		}
	}()

	log.Printf("Starting gRPC server Driver service on port %s", lis.Addr().String())

	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			log.Printf("failed to serve: %v", err)
			cancel()
		}
	}()

	// wait for the shutdown signal
	<-ctx.Done()
	log.Println("Shutting down the server...")
	grpcServer.GracefulStop()
}
