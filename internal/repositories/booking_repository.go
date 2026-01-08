package repositories

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"airline-booking-system/internal/models"
	"airline-booking-system/pkg/database"
)

// BookingRepository handles booking database operations
type BookingRepository struct {
	db *database.DB
}

// NewBookingRepository creates a new booking repository
func NewBookingRepository(db *database.DB) *BookingRepository {
	return &BookingRepository{db: db}
}

// CreateBooking creates a new booking
func (r *BookingRepository) CreateBooking(ctx context.Context, booking *models.Booking) (*models.Booking, error) {
	metadataJSON, err := json.Marshal(booking.BookingMetadata)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal booking metadata: %w", err)
	}

	query := `
		INSERT INTO bookings (flight_id, user_id, status, payment_reference_id, 
		                     booking_price, seats_booked, booking_metadata, 
		                     created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id
	`

	now := time.Now()
	err = r.db.QueryRowContext(ctx, query,
		booking.FlightID, booking.UserID, booking.Status, booking.PaymentReferenceID,
		booking.BookingPrice, booking.SeatsBooked, string(metadataJSON), now, now,
	).Scan(&booking.ID)

	if err != nil {
		return nil, fmt.Errorf("failed to create booking: %w", err)
	}

	booking.CreatedAt = now
	booking.UpdatedAt = now

	return booking, nil
}

// GetBookingByID gets a booking by ID
func (r *BookingRepository) GetBookingByID(ctx context.Context, id int64) (*models.Booking, error) {
	query := `
		SELECT id, flight_id, user_id, status, payment_reference_id, 
		       booking_price, seats_booked, booking_metadata, created_at, updated_at
		FROM bookings
		WHERE id = $1
	`

	var booking models.Booking
	var metadataJSON string

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&booking.ID, &booking.FlightID, &booking.UserID, &booking.Status,
		&booking.PaymentReferenceID, &booking.BookingPrice, &booking.SeatsBooked,
		&metadataJSON, &booking.CreatedAt, &booking.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("booking not found")
		}
		return nil, fmt.Errorf("failed to get booking: %w", err)
	}

	// Unmarshal booking metadata
	err = json.Unmarshal([]byte(metadataJSON), &booking.BookingMetadata)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal booking metadata: %w", err)
	}

	return &booking, nil
}

// UpdateBookingStatus updates the status of a booking
func (r *BookingRepository) UpdateBookingStatus(ctx context.Context, bookingID int64, status models.BookingStatus, paymentRefID *string) error {
	query := `
		UPDATE bookings 
		SET status = $1, payment_reference_id = $2, updated_at = $3
		WHERE id = $4
	`

	var paymentRef interface{}
	if paymentRefID != nil {
		paymentRef = *paymentRefID
	}

	result, err := r.db.ExecContext(ctx, query, status, paymentRef, time.Now(), bookingID)
	if err != nil {
		return fmt.Errorf("failed to update booking status: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("booking not found")
	}

	return nil
}

// GetBookingsByUserID gets bookings for a user
func (r *BookingRepository) GetBookingsByUserID(ctx context.Context, userID int64) ([]models.Booking, error) {
	query := `
		SELECT id, flight_id, user_id, status, payment_reference_id, 
		       booking_price, seats_booked, booking_metadata, created_at, updated_at
		FROM bookings
		WHERE user_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user bookings: %w", err)
	}
	defer rows.Close()

	var bookings []models.Booking
	for rows.Next() {
		var booking models.Booking
		var metadataJSON string

		err := rows.Scan(
			&booking.ID, &booking.FlightID, &booking.UserID, &booking.Status,
			&booking.PaymentReferenceID, &booking.BookingPrice, &booking.SeatsBooked,
			&metadataJSON, &booking.CreatedAt, &booking.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan booking: %w", err)
		}

		// Unmarshal booking metadata
		err = json.Unmarshal([]byte(metadataJSON), &booking.BookingMetadata)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal booking metadata: %w", err)
		}

		bookings = append(bookings, booking)
	}

	return bookings, rows.Err()
}

// GetBookingsByFlightID gets bookings for a specific flight
func (r *BookingRepository) GetBookingsByFlightID(ctx context.Context, flightID int64) ([]models.Booking, error) {
	query := `
		SELECT id, flight_id, user_id, status, payment_reference_id, 
		       booking_price, seats_booked, booking_metadata, created_at, updated_at
		FROM bookings
		WHERE flight_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, flightID)
	if err != nil {
		return nil, fmt.Errorf("failed to get flight bookings: %w", err)
	}
	defer rows.Close()

	var bookings []models.Booking
	for rows.Next() {
		var booking models.Booking
		var metadataJSON string

		err := rows.Scan(
			&booking.ID, &booking.FlightID, &booking.UserID, &booking.Status,
			&booking.PaymentReferenceID, &booking.BookingPrice, &booking.SeatsBooked,
			&metadataJSON, &booking.CreatedAt, &booking.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan booking: %w", err)
		}

		// Unmarshal booking metadata
		err = json.Unmarshal([]byte(metadataJSON), &booking.BookingMetadata)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal booking metadata: %w", err)
		}

		bookings = append(bookings, booking)
	}

	return bookings, rows.Err()
}
