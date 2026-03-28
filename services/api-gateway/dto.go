package main

import (
	pb "ride-sharing/shared/proto/trip"
	"ride-sharing/shared/types"
)

type previewTripRequest struct {
	UserID      string           `json:"userID"`
	Pickup      types.Coordinate `json:"pickup"`
	Destination types.Coordinate `json:"destination"`
}

func (r *previewTripRequest) ToProto() *pb.PreviewTripRequest {
	return &pb.PreviewTripRequest{
		UserId: r.UserID,
		StartLocation: &pb.Coordinate{
			Latitude:  r.Pickup.Latitude,
			Longitude: r.Pickup.Longitude,
		},
		EndLocation: &pb.Coordinate{
			Latitude:  r.Destination.Latitude,
			Longitude: r.Destination.Longitude,
		},
	}
}
