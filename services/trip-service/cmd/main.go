package main

import (
	"context"
	"ride-sharing/services/trip-service/internal/domain"
	"ride-sharing/services/trip-service/internal/infrastructure/repository"
	"ride-sharing/services/trip-service/internal/service"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

func main() {
	ctx := context.Background()

	inmemoryRepo := repository.NewInMemoryTripRepository()
	service := service.NewTripService(inmemoryRepo)

	// Example usage
	fare := &domain.RideFareModel{
		ID:                primitive.NewObjectID(),
		UserID:            "user123",
		PackageSlug:       "van",
		TotalPriceInCents: 150.8,
		ExpiresAt:         time.Now().Add(30 * time.Minute),
	}
	trip, err := service.CreateTrip(ctx, fare)
	if err != nil {
		panic(err)
	}
	println("Created trip with ID:", trip.ID.Hex())

	// make it run forever
	for {
		time.Sleep(3 * time.Second)
	}
}
