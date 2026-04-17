package domain

import (
	"time"

	"ride-sharing/services/trip-service/pkg/types"
	pb "ride-sharing/shared/proto/trip"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type RideFareModel struct {
	ID                primitive.ObjectID     `json:"_id" bson:"_id,omitempty"`
	UserID            string                 `json:"user_id" bson:"userID"`
	PackageSlug       string                 `json:"package_slug" bson:"packageSlug"` // van, luxury, etc
	TotalPriceInCents float64                `json:"total_price_in_cents" bson:"totalPriceInCents"`
	ExpiresAt         time.Time              `json:"expires_at" bson:"expiresAt"`
	Route             *types.OsrmApiResponse `json:"route" bson:"route"`
}

func (r *RideFareModel) ToProto() *pb.RideFare {
	return &pb.RideFare{
		Id:                r.ID.Hex(),
		UserID:            r.UserID,
		PackageSlug:       r.PackageSlug,
		TotalPriceInCents: r.TotalPriceInCents,
		// Route:             r.Route.ToProto(),
	}
}

func ToRideFaresProto(fares []*RideFareModel) []*pb.RideFare {
	protoFares := make([]*pb.RideFare, len(fares))
	for i, fare := range fares {
		protoFares[i] = fare.ToProto()
	}
	return protoFares
}
