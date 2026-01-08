package cache

import (
	"context"
	"encoding/json"
	"fmt"

	"airline-booking-system/internal/config"
	"airline-booking-system/internal/models"
	"airline-booking-system/pkg/redis"
)

// FlightCacheService handles flight caching operations
type FlightCacheService struct {
	redisClient *redis.Client
	config      *config.AppConfig
}

// NewFlightCacheService creates a new flight cache service
func NewFlightCacheService(redisClient *redis.Client, config *config.AppConfig) *FlightCacheService {
	return &FlightCacheService{
		redisClient: redisClient,
		config:      config,
	}
}

// GetCachedFlights gets flights from cache
func (s *FlightCacheService) GetCachedFlights(ctx context.Context, cacheKey string) ([]models.Flight, error) {
	cachedData, err := s.redisClient.Get(ctx, cacheKey)
	if err != nil {
		return nil, err
	}

	var flights []models.Flight
	err = json.Unmarshal([]byte(cachedData), &flights)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal cached flights: %w", err)
	}

	return flights, nil
}

// SetCachedFlights sets flights in cache
func (s *FlightCacheService) SetCachedFlights(ctx context.Context, cacheKey string, flights []models.Flight) error {
	flightData, err := json.Marshal(flights)
	if err != nil {
		return fmt.Errorf("failed to marshal flights for cache: %w", err)
	}

	return s.redisClient.SetJSON(ctx, cacheKey, string(flightData), s.config.CacheTTL)
}

// IsCached checks if a search is cached
func (s *FlightCacheService) IsCached(ctx context.Context, cacheKey string) (bool, error) {
	return s.redisClient.Exists(ctx, cacheKey)
}

// AcquireFlightLock acquires a distributed lock for a flight
func (s *FlightCacheService) AcquireFlightLock(ctx context.Context, lockKey string) (bool, error) {
	return s.redisClient.AcquireLock(ctx, lockKey, s.config.LockTTL)
}

// ReleaseFlightLock releases a distributed lock for a flight
func (s *FlightCacheService) ReleaseFlightLock(ctx context.Context, lockKey string) error {
	return s.redisClient.ReleaseLock(ctx, lockKey)
}

// GetAvailableSeats gets available seats for a flight from cache
func (s *FlightCacheService) GetAvailableSeats(ctx context.Context, flightID int64) (int, error) {
	key := fmt.Sprintf("flight_seats:%d", flightID)
	seats, err := s.redisClient.GetInt(ctx, key)
	if err != nil {
		return 0, err
	}
	return int(seats), nil
}

// SetAvailableSeats sets available seats for a flight in cache
func (s *FlightCacheService) SetAvailableSeats(ctx context.Context, flightID int64, seats int) error {
	key := fmt.Sprintf("flight_seats:%d", flightID)
	return s.redisClient.SetJSON(ctx, key, seats, s.config.CacheTTL)
}

// DecrementAvailableSeats decrements available seats for a flight
func (s *FlightCacheService) DecrementAvailableSeats(ctx context.Context, flightID int64, decrement int) error {
	key := fmt.Sprintf("flight_seats:%d", flightID)
	_, err := s.redisClient.IncrBy(ctx, key, -int64(decrement))
	return err
}

// DeleteCachedSeats removes cached seat information
func (s *FlightCacheService) DeleteCachedSeats(ctx context.Context, flightID int64) error {
	key := fmt.Sprintf("flight_seats:%d", flightID)
	return s.redisClient.Delete(ctx, key)
}
