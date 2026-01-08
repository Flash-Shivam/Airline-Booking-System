package models

import (
	"time"
)

// FlightStatus represents the status of a flight
type FlightStatus string

const (
	FlightStatusScheduled FlightStatus = "scheduled"
	FlightStatusOnTime    FlightStatus = "on_time"
	FlightStatusDelayed   FlightStatus = "delayed"
	FlightStatusDeparted  FlightStatus = "departed"
	FlightStatusCancelled FlightStatus = "cancelled"
)

// Flight represents a flight entity
type Flight struct {
	ID             int64         `json:"id" db:"id"`
	Source         string        `json:"source" db:"source"`
	Destination    string        `json:"destination" db:"destination"`
	Timestamp      time.Time     `json:"timestamp" db:"timestamp"`
	AvailableSeats int           `json:"available_seats" db:"available_seats"`
	TotalSeats     int           `json:"total_seats" db:"total_seats"`
	FlightStatus   FlightStatus  `json:"flight_status" db:"flight_status"`
	Price          float64       `json:"price" db:"price"`
	Version        int           `json:"version" db:"version"`
	CreatedAt      time.Time     `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time     `json:"updated_at" db:"updated_at"`
}

// FlightSearchRequest represents search parameters for flights
type FlightSearchRequest struct {
	Source      string    `json:"source"`
	Destination string    `json:"destination"`
	Date        time.Time `json:"date"`
}

// FlightSearchResponse represents the response for flight search
type FlightSearchResponse struct {
	Flights []Flight `json:"flights"`
	Count   int      `json:"count"`
}

// IsValid checks if the flight search request is valid
func (fsr *FlightSearchRequest) IsValid() bool {
	return fsr.Source != "" && fsr.Destination != "" && !fsr.Date.IsZero()
}

// GetCacheKey returns the Redis cache key for this search
func (fsr *FlightSearchRequest) GetCacheKey() string {
	return fsr.Source + "#" + fsr.Destination + "#" + fsr.Date.Format("2006-01-02")
}
