package services

import (
	"context"
	"fmt"
	"log"

	"airline-booking-system/internal/cache"
	"airline-booking-system/internal/config"
	"airline-booking-system/internal/models"
	"airline-booking-system/internal/repositories"
)

// FlightRepository defines the persistence operations used by FlightService.
type FlightRepository interface {
	SearchFlights(ctx context.Context, req *models.FlightSearchRequest) ([]models.Flight, error)
	GetFlightByID(ctx context.Context, id int64) (*models.Flight, error)
	CreateFlight(ctx context.Context, flight *models.Flight) (*models.Flight, error)
	UpdateFlight(ctx context.Context, flight *models.Flight) error
}

// FlightCache defines the caching operations used by FlightService.
type FlightCache interface {
	GetCachedFlights(ctx context.Context, key string) ([]models.Flight, error)
	SetCachedFlights(ctx context.Context, key string, flights []models.Flight) error
}

// FlightService handles flight business logic
type FlightService struct {
	flightRepo   FlightRepository
	cacheService FlightCache
	config       *config.AppConfig
}

// NewFlightService creates a new flight service
func NewFlightService(flightRepo *repositories.FlightRepository, cacheService *cache.FlightCacheService, config *config.AppConfig) *FlightService {
	return &FlightService{
		flightRepo:   flightRepo,
		cacheService: cacheService,
		config:       config,
	}
}

// SearchFlights searches for flights with caching
func (s *FlightService) SearchFlights(ctx context.Context, req *models.FlightSearchRequest) (*models.FlightSearchResponse, error) {
	if !req.IsValid() {
		return nil, fmt.Errorf("invalid search request")
	}

	cacheKey := req.GetCacheKey()

	// Try to get from cache first
	if flights, err := s.cacheService.GetCachedFlights(ctx, cacheKey); err == nil {
		log.Printf("Cache hit for search: %s", cacheKey)
		return &models.FlightSearchResponse{
			Flights: flights,
			Count:   len(flights),
		}, nil
	}

	// Cache miss - query database
	log.Printf("Cache miss for search: %s, querying database", cacheKey)
	flights, err := s.flightRepo.SearchFlights(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to search flights: %w", err)
	}

	// Cache the results
	if err := s.cacheService.SetCachedFlights(ctx, cacheKey, flights); err != nil {
		log.Printf("Failed to cache search results: %v", err)
		// Don't fail the request if caching fails
	}

	return &models.FlightSearchResponse{
		Flights: flights,
		Count:   len(flights),
	}, nil
}

// GetFlightByID gets a flight by ID
func (s *FlightService) GetFlightByID(ctx context.Context, id int64) (*models.Flight, error) {
	return s.flightRepo.GetFlightByID(ctx, id)
}

// CreateFlight creates a new flight
func (s *FlightService) CreateFlight(ctx context.Context, flight *models.Flight) (*models.Flight, error) {
	// Validate flight data
	if flight.Source == "" || flight.Destination == "" || flight.AvailableSeats <= 0 || flight.TotalSeats <= 0 || flight.Price <= 0 {
		return nil, fmt.Errorf("invalid flight data")
	}

	if flight.AvailableSeats > flight.TotalSeats {
		return nil, fmt.Errorf("available seats cannot exceed total seats")
	}

	if flight.Source == flight.Destination {
		return nil, fmt.Errorf("source and destination cannot be the same")
	}

	// Set default status if not provided
	if flight.FlightStatus == "" {
		flight.FlightStatus = models.FlightStatusScheduled
	}

	flight.Version = 1

	return s.flightRepo.CreateFlight(ctx, flight)
}

// UpdateFlight updates an existing flight
func (s *FlightService) UpdateFlight(ctx context.Context, flight *models.Flight) error {
	// Validate flight data
	if flight.Source == "" || flight.Destination == "" || flight.TotalSeats <= 0 || flight.Price <= 0 {
		return fmt.Errorf("invalid flight data")
	}

	if flight.AvailableSeats > flight.TotalSeats {
		return fmt.Errorf("available seats cannot exceed total seats")
	}

	if flight.Source == flight.Destination {
		return fmt.Errorf("source and destination cannot be the same")
	}

	return s.flightRepo.UpdateFlight(ctx, flight)
}
