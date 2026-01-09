package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	"airline-booking-system/internal/models"

	"github.com/gorilla/mux"
)

// BookingService defines the interface for booking-related business logic.
// This allows the HTTP handlers to be unit tested with mocks.
type BookingService interface {
	CreateBooking(rctx context.Context, req *models.BookingRequest) (*models.BookingResponse, error)
	GetBookingByID(rctx context.Context, id int64) (*models.Booking, error)
	GetBookingsByUserID(rctx context.Context, userID int64) ([]models.Booking, error)
}

// BookingHandler handles booking-related HTTP requests
type BookingHandler struct {
	bookingService BookingService
}

// NewBookingHandler creates a new booking handler
func NewBookingHandler(bookingService BookingService) *BookingHandler {
	return &BookingHandler{
		bookingService: bookingService,
	}
}

// CreateBooking handles booking creation requests
func (h *BookingHandler) CreateBooking(w http.ResponseWriter, r *http.Request) {
	var req models.BookingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}

	response, err := h.bookingService.CreateBooking(r.Context(), &req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// GetBooking handles getting a booking by ID
func (h *BookingHandler) GetBooking(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid booking ID", http.StatusBadRequest)
		return
	}

	booking, err := h.bookingService.GetBookingByID(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(booking)
}

// GetUserBookings handles getting bookings for a user
func (h *BookingHandler) GetUserBookings(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userIDStr := vars["userId"]

	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	bookings, err := h.bookingService.GetBookingsByUserID(r.Context(), userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"bookings": bookings,
		"count":    len(bookings),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
