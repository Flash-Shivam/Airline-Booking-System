package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"airline-booking-system/internal/models"

	"github.com/gorilla/mux"
)

// mockFlightService is a test double for FlightService.
type mockFlightService struct {
	searchResp *models.FlightSearchResponse
	searchErr  error

	getFlightResp *models.Flight
	getFlightErr  error

	createResp *models.Flight
	createErr  error

	updateErr error
}

func (m *mockFlightService) SearchFlights(ctx context.Context, req *models.FlightSearchRequest) (*models.FlightSearchResponse, error) {
	return m.searchResp, m.searchErr
}

func (m *mockFlightService) GetFlightByID(ctx context.Context, id int64) (*models.Flight, error) {
	return m.getFlightResp, m.getFlightErr
}

func (m *mockFlightService) CreateFlight(ctx context.Context, flight *models.Flight) (*models.Flight, error) {
	return m.createResp, m.createErr
}

func (m *mockFlightService) UpdateFlight(ctx context.Context, flight *models.Flight) error {
	return m.updateErr
}

func TestSearchFlights_Success(t *testing.T) {
	service := &mockFlightService{
		searchResp: &models.FlightSearchResponse{
			Flights: []models.Flight{
				{
					ID:            1,
					Source:        "Delhi",
					Destination:   "Mumbai",
					Timestamp:     time.Now(),
					AvailableSeats: 100,
					TotalSeats:    150,
					Price:         2500,
				},
			},
			Count: 1,
		},
	}

	handler := NewFlightHandler(service)

	req := httptest.NewRequest(http.MethodGet, "/flights/search?source=Delhi&destination=Mumbai&date=2025-01-20", nil)
	rr := httptest.NewRecorder()

	handler.SearchFlights(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, status)
	}

	var resp models.FlightSearchResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if resp.Count != 1 {
		t.Fatalf("expected count 1, got %d", resp.Count)
	}
}

func TestSearchFlights_MissingParams(t *testing.T) {
	service := &mockFlightService{}
	handler := NewFlightHandler(service)

	req := httptest.NewRequest(http.MethodGet, "/flights/search?source=Delhi&destination=Mumbai", nil)
	rr := httptest.NewRecorder()

	handler.SearchFlights(rr, req)

	if status := rr.Code; status != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, status)
	}
}

func TestSearchFlights_InvalidDate(t *testing.T) {
	service := &mockFlightService{}
	handler := NewFlightHandler(service)

	req := httptest.NewRequest(http.MethodGet, "/flights/search?source=Delhi&destination=Mumbai&date=invalid", nil)
	rr := httptest.NewRecorder()

	handler.SearchFlights(rr, req)

	if status := rr.Code; status != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, status)
	}
}

func TestGetFlight_Success(t *testing.T) {
	service := &mockFlightService{
		getFlightResp: &models.Flight{ID: 1},
	}
	handler := NewFlightHandler(service)

	req := httptest.NewRequest(http.MethodGet, "/flights/1", nil)
	req = mux.SetURLVars(req, map[string]string{"id": "1"})
	rr := httptest.NewRecorder()

	handler.GetFlight(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, status)
	}
}

func TestGetFlight_InvalidID(t *testing.T) {
	service := &mockFlightService{}
	handler := NewFlightHandler(service)

	req := httptest.NewRequest(http.MethodGet, "/flights/abc", nil)
	req = mux.SetURLVars(req, map[string]string{"id": "abc"})
	rr := httptest.NewRecorder()

	handler.GetFlight(rr, req)

	if status := rr.Code; status != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, status)
	}
}

func TestCreateFlight_InvalidJSON(t *testing.T) {
	service := &mockFlightService{}
	handler := NewFlightHandler(service)

	req := httptest.NewRequest(http.MethodPost, "/flights", bytes.NewBufferString(`invalid-json`))
	rr := httptest.NewRecorder()

	handler.CreateFlight(rr, req)

	if status := rr.Code; status != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, status)
	}
}

func TestCreateFlight_Success(t *testing.T) {
	service := &mockFlightService{
		createResp: &models.Flight{
			ID:          1,
			Source:      "Delhi",
			Destination: "Mumbai",
		},
	}
	handler := NewFlightHandler(service)

	body := `{
		"source": "Delhi",
		"destination": "Mumbai",
		"timestamp": "2025-01-20T10:00:00Z",
		"available_seats": 150,
		"total_seats": 180,
		"price": 2500
	}`
	req := httptest.NewRequest(http.MethodPost, "/flights", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	handler.CreateFlight(rr, req)

	if status := rr.Code; status != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, status)
	}
}

func TestUpdateFlight_InvalidID(t *testing.T) {
	service := &mockFlightService{}
	handler := NewFlightHandler(service)

	body := `{}` // body won't be parsed due to invalid id
	req := httptest.NewRequest(http.MethodPut, "/flights/abc", bytes.NewBufferString(body))
	req = mux.SetURLVars(req, map[string]string{"id": "abc"})
	rr := httptest.NewRecorder()

	handler.UpdateFlight(rr, req)

	if status := rr.Code; status != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, status)
	}
}

func TestUpdateFlight_InvalidJSON(t *testing.T) {
	service := &mockFlightService{}
	handler := NewFlightHandler(service)

	req := httptest.NewRequest(http.MethodPut, "/flights/1", bytes.NewBufferString(`invalid-json`))
	req = mux.SetURLVars(req, map[string]string{"id": "1"})
	rr := httptest.NewRecorder()

	handler.UpdateFlight(rr, req)

	if status := rr.Code; status != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, status)
	}
}


