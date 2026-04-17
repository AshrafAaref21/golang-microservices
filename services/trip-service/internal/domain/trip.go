package domain

import (
	"context"
	"ride-sharing/services/trip-service/pkg/types"
	pbd "ride-sharing/shared/proto/driver"
	pb "ride-sharing/shared/proto/trip"
	shared_types "ride-sharing/shared/types"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type TripModel struct {
	ID       primitive.ObjectID `json:"_id" bson:"_id,omitempty"`
	UserID   string             `json:"user_id" bson:"userID"`
	Status   string             `json:"status" bson:"status"`
	RideFare *RideFareModel     `json:"ride_fare" bson:"rideFare"`
	Driver   *pb.TripDriver     `json:"driver" bson:"driver"`
}

func (t *TripModel) ToProto() *pb.Trip {
	return &pb.Trip{
		Id:           t.ID.Hex(),
		UserID:       t.UserID,
		SelectedFare: t.RideFare.ToProto(),
		Status:       t.Status,
		Driver:       t.Driver,
		Route:        t.RideFare.Route.ToProto(),
	}
}

type TripRepository interface {
	CreateTrip(ctx context.Context, trip *TripModel) (*TripModel, error)
	SaveRideFare(ctx context.Context, fare *RideFareModel) error
	GetRideFareByID(ctx context.Context, id string) (*RideFareModel, error)
	GetTripByID(ctx context.Context, id string) (*TripModel, error)
	UpdateTrip(ctx context.Context, tripID string, status string, driver *pbd.Driver) error
}

type TripService interface {
	CreateTrip(ctx context.Context, fare *RideFareModel) (*TripModel, error)
	GetRoute(ctx context.Context, pickup, destination *shared_types.Coordinate, useOSRMApi bool) (*types.OsrmApiResponse, error)
	EstimatePackagesPriceWithRoute(route *types.OsrmApiResponse) []*RideFareModel
	GenerateTripFares(ctx context.Context, fares []*RideFareModel, userID string, route *types.OsrmApiResponse) ([]*RideFareModel, error)
	GetAndValidateFare(ctx context.Context, fareID, userID string) (*RideFareModel, error)
	GetTripByID(ctx context.Context, id string) (*TripModel, error)
	UpdateTrip(ctx context.Context, tripID string, status string, driver *pbd.Driver) error
}
