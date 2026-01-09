package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"airline-booking-system/internal/handlers"
	"airline-booking-system/internal/models"
)

// dummy implementations to satisfy handler constructors for router tests.
type dummyFlightService struct{}

func (d *dummyFlightService) SearchFlights(ctx context.Context, req *models.FlightSearchRequest) (*models.FlightSearchResponse, error) {
	return nil, nil
}

func (d *dummyFlightService) GetFlightByID(ctx context.Context, id int64) (*models.Flight, error) {
	return nil, nil
}

func (d *dummyFlightService) CreateFlight(ctx context.Context, flight *models.Flight) (*models.Flight, error) {
	return nil, nil
}

func (d *dummyFlightService) UpdateFlight(ctx context.Context, flight *models.Flight) error {
	return nil
}

type dummyBookingService struct{}

func (d *dummyBookingService) CreateBooking(ctx context.Context, req *models.BookingRequest) (*models.BookingResponse, error) {
	return nil, nil
}

func (d *dummyBookingService) GetBookingByID(ctx context.Context, id int64) (*models.Booking, error) {
	return nil, nil
}

func (d *dummyBookingService) GetBookingsByUserID(ctx context.Context, userID int64) ([]models.Booking, error) {
	return nil, nil
}

func TestHealthEndpoint(t *testing.T) {
	flightHandler := handlers.NewFlightHandler(&dummyFlightService{})
	bookingHandler := handlers.NewBookingHandler(&dummyBookingService{})

	router := setupRoutes(flightHandler, bookingHandler)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/health", nil)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, status)
	}
}



