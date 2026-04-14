package repository

import (
	"context"
	"fmt"
	"ride-sharing/services/trip-service/internal/domain"
	"sync"

	pbd "ride-sharing/shared/proto/driver"
	pb "ride-sharing/shared/proto/trip"
)

type inMemoryRepository struct {
	mu        sync.RWMutex
	trips     map[string]*domain.TripModel
	rideFares map[string]*domain.RideFareModel
}

func NewinMemoryRepository() *inMemoryRepository {
	return &inMemoryRepository{
		trips:     make(map[string]*domain.TripModel),
		rideFares: make(map[string]*domain.RideFareModel),
	}
}

func (r *inMemoryRepository) CreateTrip(ctx context.Context, trip *domain.TripModel) (*domain.TripModel, error) {
	r.mu.Lock()
	r.trips[trip.ID.Hex()] = trip
	r.mu.Unlock()
	return trip, nil
}

func (r *inMemoryRepository) SaveRideFare(ctx context.Context, fare *domain.RideFareModel) error {
	r.mu.Lock()
	r.rideFares[fare.ID.Hex()] = fare
	r.mu.Unlock()
	return nil
}

func (r *inMemoryRepository) GetRideFareByID(ctx context.Context, id string) (*domain.RideFareModel, error) {
	fare, exist := r.rideFares[id]
	if !exist {
		return nil, fmt.Errorf("fare does not exist with ID: %s", id)
	}

	return fare, nil
}

func (r *inMemoryRepository) GetTripByID(ctx context.Context, id string) (*domain.TripModel, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	trip, ok := r.trips[id]
	if !ok {
		return nil, nil
	}
	return trip, nil
}

func (r *inMemoryRepository) UpdateTrip(ctx context.Context, tripID string, status string, driver *pbd.Driver) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	trip, ok := r.trips[tripID]
	if !ok {
		return fmt.Errorf("trip not found with ID: %s", tripID)
	}

	trip.Status = status

	if driver != nil {
		trip.Driver = &pb.TripDriver{
			Id:             driver.Id,
			Name:           driver.Name,
			CarPlate:       driver.CarPlate,
			ProfilePicture: driver.ProfilePicture,
		}
	}
	return nil
}
