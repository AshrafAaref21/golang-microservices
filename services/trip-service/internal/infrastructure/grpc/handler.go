package grpc

import (
	"context"
	"log"
	"ride-sharing/services/trip-service/internal/domain"
	"ride-sharing/services/trip-service/internal/infrastructure/events"
	pb "ride-sharing/shared/proto/trip"
	"ride-sharing/shared/types"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type grpcHandler struct {
	pb.UnimplementedTripServiceServer
	service        domain.TripService
	eventPublisher *events.TripEventPublisher
}

func NewGrpcHandler(server *grpc.Server, service domain.TripService, eventPublisher *events.TripEventPublisher) *grpcHandler {
	handler := &grpcHandler{service: service, eventPublisher: eventPublisher}
	pb.RegisterTripServiceServer(server, handler)
	return handler
}

func (h *grpcHandler) PreviewTrip(ctx context.Context, req *pb.PreviewTripRequest) (*pb.PreviewTripResponse, error) {
	pickup := req.GetStartLocation()
	destination := req.GetEndLocation()

	pickupCoord := &types.Coordinate{
		Latitude:  pickup.GetLatitude(),
		Longitude: pickup.GetLongitude(),
	}
	destinationCoord := &types.Coordinate{
		Latitude:  destination.GetLatitude(),
		Longitude: destination.GetLongitude(),
	}

	t, err := h.service.GetRoute(ctx, pickupCoord, destinationCoord, false) // Set to false to use mock response for development
	if err != nil {
		log.Printf("Error getting route: %v", err)
		return nil, status.Error(codes.Internal, err.Error())
	}

	estimatedFares := h.service.EstimatePackagesPriceWithRoute(t)
	log.Printf("Estimated fares: %+v", estimatedFares)

	fares, err := h.service.GenerateTripFares(ctx, estimatedFares, req.GetUserID(), t)
	if err != nil {
		log.Printf("Error generating trip fares: %v", err)
		return nil, status.Error(codes.Internal, "failed to generate trip fares")
	}

	return &pb.PreviewTripResponse{
		Route:     t.ToProto(),
		RideFares: domain.ToRideFaresProto(fares),
	}, nil
}

func (h *grpcHandler) CreateTrip(ctx context.Context, req *pb.CreateTripRequest) (*pb.CreateTripResponse, error) {
	fareID := req.GetRideFareID()
	userID := req.GetUserID()

	rideFare, err := h.service.GetAndValidateFare(ctx, fareID, userID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to validate the fare: %v", err)
	}

	trip, err := h.service.CreateTrip(ctx, rideFare)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create the trip: %v", err)
	}

	// Publish an event on the Async Comms module.
	if err := h.eventPublisher.PublishTripCreated(ctx, trip); err != nil {
		log.Printf("Error publishing trip created event: %v", err)
	}

	return &pb.CreateTripResponse{
		TripID: trip.ID.Hex(),
	}, nil
}
