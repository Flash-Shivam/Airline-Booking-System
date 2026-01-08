package models

import (
	"time"
)

// BookingStatus represents the status of a booking
type BookingStatus string

const (
	BookingStatusPending   BookingStatus = "pending"
	BookingStatusCompleted BookingStatus = "completed"
	BookingStatusFailed    BookingStatus = "failed"
	BookingStatusCancelled BookingStatus = "cancelled"
)

// PassengerDetails represents passenger information
type PassengerDetails struct {
	Name    string `json:"name"`
	Email   string `json:"email"`
	Phone   string `json:"phone"`
	Age     int    `json:"age"`
	Gender  string `json:"gender"`
}

// Booking represents a booking entity
type Booking struct {
	ID                int64             `json:"id" db:"id"`
	FlightID          int64             `json:"flight_id" db:"flight_id"`
	UserID            int64             `json:"user_id" db:"user_id"`
	Status            BookingStatus     `json:"status" db:"status"`
	PaymentReferenceID string           `json:"payment_reference_id" db:"payment_reference_id"`
	BookingPrice      float64           `json:"booking_price" db:"booking_price"`
	SeatsBooked       int               `json:"seats_booked" db:"seats_booked"`
	BookingMetadata   []PassengerDetails `json:"booking_metadata" db:"booking_metadata"`
	CreatedAt         time.Time         `json:"created_at" db:"created_at"`
	UpdatedAt         time.Time         `json:"updated_at" db:"updated_at"`
}

// BookingRequest represents a booking creation request
type BookingRequest struct {
	FlightID        int64             `json:"flight_id"`
	UserID          int64             `json:"user_id"`
	SeatsBooked     int               `json:"seats_booked"`
	PassengerDetails []PassengerDetails `json:"passenger_details"`
}

// BookingResponse represents the response for booking operations
type BookingResponse struct {
	BookingID         int64         `json:"booking_id"`
	Status           BookingStatus `json:"status"`
	PaymentReferenceID string       `json:"payment_reference_id,omitempty"`
	Message          string        `json:"message"`
}

// SeatUpdateEvent represents an event for seat updates
type SeatUpdateEvent struct {
	FlightID     int64     `json:"flight_id"`
	SeatsBooked  int       `json:"seats_booked"`
	Timestamp    time.Time `json:"timestamp"`
	BookingID    int64     `json:"booking_id"`
}

// PaymentEvent represents a payment processing event
type PaymentEvent struct {
	BookingID         int64     `json:"booking_id"`
	PaymentReferenceID string   `json:"payment_reference_id"`
	Amount           float64    `json:"amount"`
	Status           string     `json:"status"`
	Timestamp        time.Time  `json:"timestamp"`
}

// IsValid checks if the booking request is valid
func (br *BookingRequest) IsValid() bool {
	return br.FlightID > 0 && br.UserID > 0 && br.SeatsBooked > 0 && len(br.PassengerDetails) > 0
}

// GetLockKey returns the Redis lock key for this flight
func (br *BookingRequest) GetLockKey() string {
	return "flight_lock:" + string(rune(br.FlightID))
}
