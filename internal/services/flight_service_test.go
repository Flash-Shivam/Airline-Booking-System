package services

import (
	"context"
	"errors"
	"testing"
	"time"

	"airline-booking-system/internal/models"
)

// mockFlightRepo implements FlightRepository for testing.
type mockFlightRepo struct {
	searchFlightsFn      func(ctx context.Context, req *models.FlightSearchRequest) ([]models.Flight, error)
	getFlightByIDFn      func(ctx context.Context, id int64) (*models.Flight, error)
	createFlightFn       func(ctx context.Context, flight *models.Flight) (*models.Flight, error)
	updateFlightFn       func(ctx context.Context, flight *models.Flight) error
	updateAvailableSeats func(ctx context.Context, flightID int64, seatsToBook int, version int) error
}

func (m *mockFlightRepo) SearchFlights(ctx context.Context, req *models.FlightSearchRequest) ([]models.Flight, error) {
	if m.searchFlightsFn != nil {
		return m.searchFlightsFn(ctx, req)
	}
	return nil, nil
}

func (m *mockFlightRepo) GetFlightByID(ctx context.Context, id int64) (*models.Flight, error) {
	if m.getFlightByIDFn != nil {
		return m.getFlightByIDFn(ctx, id)
	}
	return nil, nil
}

func (m *mockFlightRepo) CreateFlight(ctx context.Context, flight *models.Flight) (*models.Flight, error) {
	if m.createFlightFn != nil {
		return m.createFlightFn(ctx, flight)
	}
	return flight, nil
}

func (m *mockFlightRepo) UpdateFlight(ctx context.Context, flight *models.Flight) error {
	if m.updateFlightFn != nil {
		return m.updateFlightFn(ctx, flight)
	}
	return nil
}

func (m *mockFlightRepo) UpdateAvailableSeats(ctx context.Context, flightID int64, seatsToBook int, version int) error {
	if m.updateAvailableSeats != nil {
		return m.updateAvailableSeats(ctx, flightID, seatsToBook, version)
	}
	return nil
}

// mockFlightCache implements FlightCache for testing.
type mockFlightCache struct {
	getFn func(ctx context.Context, key string) ([]models.Flight, error)
	setFn func(ctx context.Context, key string, flights []models.Flight) error
}

func (m *mockFlightCache) GetCachedFlights(ctx context.Context, key string) ([]models.Flight, error) {
	if m.getFn != nil {
		return m.getFn(ctx, key)
	}
	return nil, errors.New("cache miss")
}

func (m *mockFlightCache) SetCachedFlights(ctx context.Context, key string, flights []models.Flight) error {
	if m.setFn != nil {
		return m.setFn(ctx, key, flights)
	}
	return nil
}

func TestFlightService_SearchFlights_InvalidRequest(t *testing.T) {
	repo := &mockFlightRepo{}
	cache := &mockFlightCache{}
	svc := &FlightService{flightRepo: repo, cacheService: cache}

	// Invalid because Date is zero
	req := &models.FlightSearchRequest{
		Source:      "Delhi",
		Destination: "Mumbai",
	}

	if _, err := svc.SearchFlights(context.Background(), req); err == nil {
		t.Fatalf("expected error for invalid request, got nil")
	}
}

func TestFlightService_SearchFlights_CacheHit(t *testing.T) {
	expected := []models.Flight{
		{ID: 1, Source: "Delhi", Destination: "Mumbai"},
	}

	repo := &mockFlightRepo{}
	cache := &mockFlightCache{
		getFn: func(ctx context.Context, key string) ([]models.Flight, error) {
			return expected, nil
		},
	}
	svc := &FlightService{flightRepo: repo, cacheService: cache}

	req := &models.FlightSearchRequest{
		Source:      "Delhi",
		Destination: "Mumbai",
		Date:        time.Date(2025, 1, 20, 0, 0, 0, 0, time.UTC),
	}

	resp, err := svc.SearchFlights(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp.Count != 1 {
		t.Fatalf("expected count 1, got %d", resp.Count)
	}
}

func TestFlightService_SearchFlights_CacheMiss_DBHit(t *testing.T) {
	expected := []models.Flight{
		{ID: 1, Source: "Delhi", Destination: "Mumbai"},
	}

	repo := &mockFlightRepo{
		searchFlightsFn: func(ctx context.Context, req *models.FlightSearchRequest) ([]models.Flight, error) {
			return expected, nil
		},
	}
	cache := &mockFlightCache{
		getFn: func(ctx context.Context, key string) ([]models.Flight, error) {
			return nil, errors.New("cache miss")
		},
	}

	svc := &FlightService{flightRepo: repo, cacheService: cache}

	req := &models.FlightSearchRequest{
		Source:      "Delhi",
		Destination: "Mumbai",
		Date:        time.Date(2025, 1, 20, 0, 0, 0, 0, time.UTC),
	}

	resp, err := svc.SearchFlights(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp.Count != 1 {
		t.Fatalf("expected count 1, got %d", resp.Count)
	}
}

func TestFlightService_CreateFlight_ValidationErrors(t *testing.T) {
	repo := &mockFlightRepo{}
	cache := &mockFlightCache{}
	svc := &FlightService{flightRepo: repo, cacheService: cache}

	tests := []struct {
		name   string
		flight *models.Flight
	}{
		{
			name: "missing fields",
			flight: &models.Flight{
				Source:         "",
				Destination:    "Mumbai",
				AvailableSeats: 10,
				TotalSeats:     20,
				Price:          100,
			},
		},
		{
			name: "available > total",
			flight: &models.Flight{
				Source:         "Delhi",
				Destination:    "Mumbai",
				AvailableSeats: 30,
				TotalSeats:     20,
				Price:          100,
			},
		},
		{
			name: "same source and destination",
			flight: &models.Flight{
				Source:         "Delhi",
				Destination:    "Delhi",
				AvailableSeats: 10,
				TotalSeats:     20,
				Price:          100,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := svc.CreateFlight(context.Background(), tt.flight); err == nil {
				t.Fatalf("expected error, got nil")
			}
		})
	}
}

func TestFlightService_CreateFlight_SetsDefaultsAndCallsRepo(t *testing.T) {
	called := false

	repo := &mockFlightRepo{
		createFlightFn: func(ctx context.Context, f *models.Flight) (*models.Flight, error) {
			called = true
			return f, nil
		},
	}
	cache := &mockFlightCache{}
	svc := &FlightService{flightRepo: repo, cacheService: cache}

	flight := &models.Flight{
		Source:         "Delhi",
		Destination:    "Mumbai",
		AvailableSeats: 10,
		TotalSeats:     20,
		Price:          100,
	}

	created, err := svc.CreateFlight(context.Background(), flight)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !called {
		t.Fatalf("expected repo.CreateFlight to be called")
	}

	if created.FlightStatus == "" {
		t.Fatalf("expected default flight status to be set")
	}

	if created.Version != 1 {
		t.Fatalf("expected version 1, got %d", created.Version)
	}
}

func TestFlightService_UpdateFlight_ValidationErrors(t *testing.T) {
	repo := &mockFlightRepo{}
	cache := &mockFlightCache{}
	svc := &FlightService{flightRepo: repo, cacheService: cache}

	flight := &models.Flight{
		Source:      "",
		Destination: "Mumbai",
		TotalSeats:  20,
		Price:       100,
	}

	if err := svc.UpdateFlight(context.Background(), flight); err == nil {
		t.Fatalf("expected error for invalid data, got nil")
	}
}

func TestFlightService_UpdateFlight_Success(t *testing.T) {
	called := false
	repo := &mockFlightRepo{
		updateFlightFn: func(ctx context.Context, f *models.Flight) error {
			called = true
			return nil
		},
	}
	cache := &mockFlightCache{}
	svc := &FlightService{flightRepo: repo, cacheService: cache}

	flight := &models.Flight{
		Source:         "Delhi",
		Destination:    "Mumbai",
		AvailableSeats: 10,
		TotalSeats:     20,
		Price:          100,
	}

	if err := svc.UpdateFlight(context.Background(), flight); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !called {
		t.Fatalf("expected repo.UpdateFlight to be called")
	}
}


