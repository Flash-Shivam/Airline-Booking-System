package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"airline-booking-system/internal/models"

	"github.com/gorilla/mux"
)

// FlightService defines the interface for flight-related business logic.
// This allows the HTTP handlers to be unit tested with mocks.
type FlightService interface {
	SearchFlights(rctx context.Context, req *models.FlightSearchRequest) (*models.FlightSearchResponse, error)
	GetFlightByID(rctx context.Context, id int64) (*models.Flight, error)
	CreateFlight(rctx context.Context, flight *models.Flight) (*models.Flight, error)
	UpdateFlight(rctx context.Context, flight *models.Flight) error
}

// FlightHandler handles flight-related HTTP requests.
type FlightHandler struct {
	flightService FlightService
}

// NewFlightHandler creates a new flight handler.
func NewFlightHandler(flightService FlightService) *FlightHandler {
	return &FlightHandler{
		flightService: flightService,
	}
}

// SearchFlights handles flight search requests
func (h *FlightHandler) SearchFlights(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	source := r.URL.Query().Get("source")
	destination := r.URL.Query().Get("destination")
	dateStr := r.URL.Query().Get("date")

	if source == "" || destination == "" || dateStr == "" {
		http.Error(w, "Missing required parameters: source, destination, date", http.StatusBadRequest)
		return
	}

	// Parse date
	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		http.Error(w, "Invalid date format. Use YYYY-MM-DD", http.StatusBadRequest)
		return
	}

	req := &models.FlightSearchRequest{
		Source:      source,
		Destination: destination,
		Date:        date,
	}

	response, err := h.flightService.SearchFlights(r.Context(), req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetFlight handles getting a flight by ID
func (h *FlightHandler) GetFlight(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid flight ID", http.StatusBadRequest)
		return
	}

	flight, err := h.flightService.GetFlightByID(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(flight)
}

// CreateFlight handles flight creation
func (h *FlightHandler) CreateFlight(w http.ResponseWriter, r *http.Request) {
	var flight models.Flight
	if err := json.NewDecoder(r.Body).Decode(&flight); err != nil {
		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}

	createdFlight, err := h.flightService.CreateFlight(r.Context(), &flight)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(createdFlight)
}

// UpdateFlight handles flight updates
func (h *FlightHandler) UpdateFlight(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid flight ID", http.StatusBadRequest)
		return
	}

	var flight models.Flight
	if err := json.NewDecoder(r.Body).Decode(&flight); err != nil {
		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}

	flight.ID = id
	if err := h.flightService.UpdateFlight(r.Context(), &flight); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Flight updated successfully"})
}
