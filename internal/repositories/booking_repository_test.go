package repositories

import (
	"context"
	"database/sql"
	"regexp"
	"testing"
	"time"

	"airline-booking-system/internal/models"
	"airline-booking-system/pkg/database"

	"github.com/DATA-DOG/go-sqlmock"
)

// helper to create a booking repository with sqlmock
func newMockBookingRepo(t *testing.T) (*BookingRepository, sqlmock.Sqlmock, func()) {
	t.Helper()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}

	wrapped := &database.DB{DB: db}

	cleanup := func() {
		db.Close()
	}

	return NewBookingRepository(wrapped), mock, cleanup
}

func TestBookingRepository_CreateBooking_Success(t *testing.T) {
	repo, mock, cleanup := newMockBookingRepo(t)
	defer cleanup()

	booking := &models.Booking{
		FlightID:   1,
		UserID:     123,
		Status:     models.BookingStatusPending,
		BookingPrice: 5000.0,
		SeatsBooked: 2,
		BookingMetadata: []models.PassengerDetails{
			{Name: "John Doe"},
		},
	}

	mock.ExpectQuery(regexp.QuoteMeta(`
		INSERT INTO bookings (flight_id, user_id, status, payment_reference_id, 
		                     booking_price, seats_booked, booking_metadata, 
		                     created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id
	`)).
		WithArgs(
			booking.FlightID, booking.UserID, booking.Status, booking.PaymentReferenceID,
			booking.BookingPrice, booking.SeatsBooked, sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
		).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(int64(1)))

	created, err := repo.CreateBooking(context.Background(), booking)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if created.ID != 1 {
		t.Fatalf("expected id 1, got %d", created.ID)
	}
}

func TestBookingRepository_GetBookingByID_NotFound(t *testing.T) {
	repo, mock, cleanup := newMockBookingRepo(t)
	defer cleanup()

	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT id, flight_id, user_id, status, payment_reference_id, 
		       booking_price, seats_booked, booking_metadata, created_at, updated_at
		FROM bookings
		WHERE id = $1
	`)).
		WithArgs(int64(1)).
		WillReturnError(sql.ErrNoRows)

	booking, err := repo.GetBookingByID(context.Background(), 1)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}

	if booking != nil {
		t.Fatalf("expected nil booking, got %+v", booking)
	}
}

func TestBookingRepository_UpdateBookingStatus_Success(t *testing.T) {
	repo, mock, cleanup := newMockBookingRepo(t)
	defer cleanup()

	status := models.BookingStatusCompleted
	paymentRef := "PAY-123"

	mock.ExpectExec(regexp.QuoteMeta(`
		UPDATE bookings 
		SET status = $1, payment_reference_id = $2, updated_at = $3
		WHERE id = $4
	`)).
		WithArgs(status, paymentRef, sqlmock.AnyArg(), int64(1)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.UpdateBookingStatus(context.Background(), 1, status, &paymentRef)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestBookingRepository_UpdateBookingStatus_NoRows(t *testing.T) {
	repo, mock, cleanup := newMockBookingRepo(t)
	defer cleanup()

	status := models.BookingStatusCompleted

	mock.ExpectExec(regexp.QuoteMeta(`
		UPDATE bookings 
		SET status = $1, payment_reference_id = $2, updated_at = $3
		WHERE id = $4
	`)).
		WithArgs(status, nil, sqlmock.AnyArg(), int64(1)).
		WillReturnResult(sqlmock.NewResult(0, 0))

	err := repo.UpdateBookingStatus(context.Background(), 1, status, nil)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
}

func TestBookingRepository_GetBookingsByUserID_Success(t *testing.T) {
	repo, mock, cleanup := newMockBookingRepo(t)
	defer cleanup()

	now := time.Now()
	rows := sqlmock.NewRows([]string{
		"id", "flight_id", "user_id", "status", "payment_reference_id",
		"booking_price", "seats_booked", "booking_metadata", "created_at", "updated_at",
	}).AddRow(
		int64(1), int64(1), int64(123), models.BookingStatusCompleted, "PAY-1",
		5000.0, 2, `[]`, now, now,
	)

	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT id, flight_id, user_id, status, payment_reference_id, 
		       booking_price, seats_booked, booking_metadata, created_at, updated_at
		FROM bookings
		WHERE user_id = $1
		ORDER BY created_at DESC
	`)).
		WithArgs(int64(123)).
		WillReturnRows(rows)

	bookings, err := repo.GetBookingsByUserID(context.Background(), 123)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(bookings) != 1 {
		t.Fatalf("expected 1 booking, got %d", len(bookings))
	}
}

func TestBookingRepository_GetBookingsByFlightID_Success(t *testing.T) {
	repo, mock, cleanup := newMockBookingRepo(t)
	defer cleanup()

	now := time.Now()
	rows := sqlmock.NewRows([]string{
		"id", "flight_id", "user_id", "status", "payment_reference_id",
		"booking_price", "seats_booked", "booking_metadata", "created_at", "updated_at",
	}).AddRow(
		int64(1), int64(1), int64(123), models.BookingStatusCompleted, "PAY-1",
		5000.0, 2, `[]`, now, now,
	)

	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT id, flight_id, user_id, status, payment_reference_id, 
		       booking_price, seats_booked, booking_metadata, created_at, updated_at
		FROM bookings
		WHERE flight_id = $1
		ORDER BY created_at DESC
	`)).
		WithArgs(int64(1)).
		WillReturnRows(rows)

	bookings, err := repo.GetBookingsByFlightID(context.Background(), 1)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(bookings) != 1 {
		t.Fatalf("expected 1 booking, got %d", len(bookings))
	}
}


