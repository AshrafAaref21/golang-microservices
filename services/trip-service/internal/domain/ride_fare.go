package domain

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type RideFareModel struct {
	ID                primitive.ObjectID `json:"_id"`
	UserID            string             `json:"user_id"`
	PackageSlug       string             `json:"package_slug"` // van, luxury, etc
	TotalPriceInCents float64            `json:"total_price_in_cents"`
	ExpiresAt         time.Time          `json:"expires_at"`
}
