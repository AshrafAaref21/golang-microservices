package grpc

import (
	"context"
	"log"
	"ride-sharing/services/trip-service/internal/domain"
	pb "ride-sharing/shared/proto/trip"
	"ride-sharing/shared/types"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type grpcHandler struct {
	pb.UnimplementedTripServiceServer
	service domain.TripService
}

func NewGrpcHandler(server *grpc.Server, service domain.TripService) *grpcHandler {
	handler := &grpcHandler{service: service}
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

	t, err := h.service.GetRoute(ctx, pickupCoord, destinationCoord)
	if err != nil {
		log.Printf("Error getting route: %v", err)
		return nil, status.Error(codes.Internal, "err.Error()xxxxxx")
	}

	return &pb.PreviewTripResponse{
		Route:     t.ToProto(),
		RideFares: []*pb.RideFare{},
	}, nil
}
