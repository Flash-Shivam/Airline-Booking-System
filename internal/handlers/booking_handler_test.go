package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"airline-booking-system/internal/models"

	"github.com/gorilla/mux"
)

// mockBookingService is a test double for BookingService.
type mockBookingService struct {
	createResp *models.BookingResponse
	createErr  error

	getBookingResp *models.Booking
	getBookingErr  error

	getByUserResp []models.Booking
	getByUserErr  error
}

func (m *mockBookingService) CreateBooking(ctx context.Context, req *models.BookingRequest) (*models.BookingResponse, error) {
	return m.createResp, m.createErr
}

func (m *mockBookingService) GetBookingByID(ctx context.Context, id int64) (*models.Booking, error) {
	return m.getBookingResp, m.getBookingErr
}

func (m *mockBookingService) GetBookingsByUserID(ctx context.Context, userID int64) ([]models.Booking, error) {
	return m.getByUserResp, m.getByUserErr
}

func TestCreateBooking_InvalidJSON(t *testing.T) {
	service := &mockBookingService{}
	handler := NewBookingHandler(service)

	req := httptest.NewRequest(http.MethodPost, "/bookings", bytes.NewBufferString(`invalid-json`))
	rr := httptest.NewRecorder()

	handler.CreateBooking(rr, req)

	if status := rr.Code; status != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, status)
	}
}

func TestCreateBooking_Success(t *testing.T) {
	service := &mockBookingService{
		createResp: &models.BookingResponse{
			BookingID: 1,
			Status:    models.BookingStatusPending,
			Message:   "ok",
		},
	}
	handler := NewBookingHandler(service)

	body := `{
		"flight_id": 1,
		"user_id": 123,
		"seats_booked": 2,
		"passenger_details": []
	}`
	req := httptest.NewRequest(http.MethodPost, "/bookings", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	handler.CreateBooking(rr, req)

	if status := rr.Code; status != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, status)
	}

	var resp models.BookingResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if resp.BookingID != 1 {
		t.Fatalf("expected booking id 1, got %d", resp.BookingID)
	}
}

func TestGetBooking_InvalidID(t *testing.T) {
	service := &mockBookingService{}
	handler := NewBookingHandler(service)

	req := httptest.NewRequest(http.MethodGet, "/bookings/abc", nil)
	req = mux.SetURLVars(req, map[string]string{"id": "abc"})
	rr := httptest.NewRecorder()

	handler.GetBooking(rr, req)

	if status := rr.Code; status != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, status)
	}
}

func TestGetBooking_Success(t *testing.T) {
	service := &mockBookingService{
		getBookingResp: &models.Booking{ID: 1},
	}
	handler := NewBookingHandler(service)

	req := httptest.NewRequest(http.MethodGet, "/bookings/1", nil)
	req = mux.SetURLVars(req, map[string]string{"id": "1"})
	rr := httptest.NewRecorder()

	handler.GetBooking(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, status)
	}
}

func TestGetUserBookings_InvalidUserID(t *testing.T) {
	service := &mockBookingService{}
	handler := NewBookingHandler(service)

	req := httptest.NewRequest(http.MethodGet, "/users/abc/bookings", nil)
	req = mux.SetURLVars(req, map[string]string{"userId": "abc"})
	rr := httptest.NewRecorder()

	handler.GetUserBookings(rr, req)

	if status := rr.Code; status != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, status)
	}
}

func TestGetUserBookings_Success(t *testing.T) {
	service := &mockBookingService{
		getByUserResp: []models.Booking{
			{ID: 1},
			{ID: 2},
		},
	}
	handler := NewBookingHandler(service)

	req := httptest.NewRequest(http.MethodGet, "/users/123/bookings", nil)
	req = mux.SetURLVars(req, map[string]string{"userId": "123"})
	rr := httptest.NewRecorder()

	handler.GetUserBookings(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, status)
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if count, ok := resp["count"].(float64); !ok || int(count) != 2 {
		t.Fatalf("expected count 2, got %v", resp["count"])
	}
}


