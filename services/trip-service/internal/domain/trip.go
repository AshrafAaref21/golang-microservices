package domain

import (
	"context"
	"ride-sharing/services/trip-service/pkg/types"
	pb "ride-sharing/shared/proto/trip"
	shared_types "ride-sharing/shared/types"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type TripModel struct {
	ID       primitive.ObjectID `json:"_id"`
	UserID   string             `json:"user_id"`
	Status   string             `json:"status"`
	RideFare *RideFareModel     `json:"ride_fare"`
	Driver   *pb.TripDriver
}

type TripRepository interface {
	CreateTrip(ctx context.Context, trip *TripModel) (*TripModel, error)
	SaveRideFare(ctx context.Context, fare *RideFareModel) error
	GetRideFareByID(ctx context.Context, id string) (*RideFareModel, error)
}

type TripService interface {
	CreateTrip(ctx context.Context, fare *RideFareModel) (*TripModel, error)
	GetRoute(ctx context.Context, pickup, destination *shared_types.Coordinate, useOSRMApi bool) (*types.OsrmApiResponse, error)
	EstimatePackagesPriceWithRoute(route *types.OsrmApiResponse) []*RideFareModel
	GenerateTripFares(ctx context.Context, fares []*RideFareModel, userID string, route *types.OsrmApiResponse) ([]*RideFareModel, error)
	GetAndValidateFare(ctx context.Context, fareID, userID string) (*RideFareModel, error)
}
