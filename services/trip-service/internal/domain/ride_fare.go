package domain

import (
	"time"

	pb "ride-sharing/shared/proto/trip"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type RideFareModel struct {
	ID                primitive.ObjectID `json:"_id"`
	UserID            string             `json:"user_id"`
	PackageSlug       string             `json:"package_slug"` // van, luxury, etc
	TotalPriceInCents float64            `json:"total_price_in_cents"`
	ExpiresAt         time.Time          `json:"expires_at"`
}

func (r *RideFareModel) ToProto() *pb.RideFare {
	return &pb.RideFare{
		Id:                r.ID.Hex(),
		UserID:            r.UserID,
		PackageSlug:       r.PackageSlug,
		TotalPriceInCents: r.TotalPriceInCents,
	}
}

func ToRideFaresProto(fares []*RideFareModel) []*pb.RideFare {
	protoFares := make([]*pb.RideFare, len(fares))
	for i, fare := range fares {
		protoFares[i] = fare.ToProto()
	}
	return protoFares
}
