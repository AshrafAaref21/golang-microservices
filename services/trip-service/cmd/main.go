package main

import (
	"context"
	"log"
	"net"
	"os"
	"os/signal"
	"ride-sharing/services/trip-service/internal/infrastructure/grpc"
	"ride-sharing/services/trip-service/internal/infrastructure/repository"
	"ride-sharing/services/trip-service/internal/service"
	"syscall"

	grpcserver "google.golang.org/grpc"
)

var GrpcAddr = ":9083"

func main() {

	inmemoryRepo := repository.NewInMemoryTripRepository()
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

	grpcServer := grpcserver.NewServer()
	grpc.NewGrpcHandler(grpcServer, service)

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
