package repositories

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"airline-booking-system/internal/models"
	"airline-booking-system/pkg/database"
)

// FlightRepository handles flight database operations
type FlightRepository struct {
	db *database.DB
}

// NewFlightRepository creates a new flight repository
func NewFlightRepository(db *database.DB) *FlightRepository {
	return &FlightRepository{db: db}
}

// SearchFlights searches for flights based on criteria
func (r *FlightRepository) SearchFlights(ctx context.Context, req *models.FlightSearchRequest) ([]models.Flight, error) {
	query := `
		SELECT id, source, destination, timestamp, available_seats, total_seats, 
		       flight_status, price, version, created_at, updated_at
		FROM flights
		WHERE source = $1 
		  AND destination = $2 
		  AND DATE(timestamp) = $3
		  AND available_seats > 0
		  AND flight_status IN ('scheduled', 'on_time')
		ORDER BY timestamp ASC
	`

	rows, err := r.db.QueryContext(ctx, query, req.Source, req.Destination, req.Date.Format("2006-01-02"))
	if err != nil {
		return nil, fmt.Errorf("failed to search flights: %w", err)
	}
	defer rows.Close()

	var flights []models.Flight
	for rows.Next() {
		var flight models.Flight
		err := rows.Scan(
			&flight.ID, &flight.Source, &flight.Destination, &flight.Timestamp,
			&flight.AvailableSeats, &flight.TotalSeats, &flight.FlightStatus,
			&flight.Price, &flight.Version, &flight.CreatedAt, &flight.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan flight: %w", err)
		}
		flights = append(flights, flight)
	}

	return flights, rows.Err()
}

// GetFlightByID gets a flight by ID
func (r *FlightRepository) GetFlightByID(ctx context.Context, id int64) (*models.Flight, error) {
	query := `
		SELECT id, source, destination, timestamp, available_seats, total_seats, 
		       flight_status, price, version, created_at, updated_at
		FROM flights
		WHERE id = $1
	`

	var flight models.Flight
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&flight.ID, &flight.Source, &flight.Destination, &flight.Timestamp,
		&flight.AvailableSeats, &flight.TotalSeats, &flight.FlightStatus,
		&flight.Price, &flight.Version, &flight.CreatedAt, &flight.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("flight not found")
		}
		return nil, fmt.Errorf("failed to get flight: %w", err)
	}

	return &flight, nil
}

// UpdateAvailableSeats updates available seats for a flight with optimistic locking
func (r *FlightRepository) UpdateAvailableSeats(ctx context.Context, flightID int64, seatsToBook int, version int) error {
	query := `
		UPDATE flights 
		SET available_seats = available_seats - $1, 
		    version = version + 1, 
		    updated_at = $2
		WHERE id = $3 AND version = $4 AND available_seats >= $1
	`

	result, err := r.db.ExecContext(ctx, query, seatsToBook, time.Now(), flightID, version)
	if err != nil {
		return fmt.Errorf("failed to update available seats: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("optimistic lock failed or insufficient seats")
	}

	return nil
}

// CreateFlight creates a new flight
func (r *FlightRepository) CreateFlight(ctx context.Context, flight *models.Flight) (*models.Flight, error) {
	query := `
		INSERT INTO flights (source, destination, timestamp, available_seats, total_seats, 
		                    flight_status, price, version, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id
	`

	now := time.Now()
	err := r.db.QueryRowContext(ctx, query,
		flight.Source, flight.Destination, flight.Timestamp,
		flight.AvailableSeats, flight.TotalSeats, flight.FlightStatus,
		flight.Price, flight.Version, now, now,
	).Scan(&flight.ID)

	if err != nil {
		return nil, fmt.Errorf("failed to create flight: %w", err)
	}

	flight.CreatedAt = now
	flight.UpdatedAt = now

	return flight, nil
}

// UpdateFlight updates an existing flight
func (r *FlightRepository) UpdateFlight(ctx context.Context, flight *models.Flight) error {
	query := `
		UPDATE flights 
		SET source = $1, destination = $2, timestamp = $3, available_seats = $4, 
		    total_seats = $5, flight_status = $6, price = $7, version = version + 1, 
		    updated_at = $8
		WHERE id = $9 AND version = $10
	`

	result, err := r.db.ExecContext(ctx, query,
		flight.Source, flight.Destination, flight.Timestamp, flight.AvailableSeats,
		flight.TotalSeats, flight.FlightStatus, flight.Price, time.Now(),
		flight.ID, flight.Version,
	)

	if err != nil {
		return fmt.Errorf("failed to update flight: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("flight not found or version conflict")
	}

	return nil
}
