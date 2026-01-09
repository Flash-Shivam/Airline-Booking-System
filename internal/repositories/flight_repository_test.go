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

// helper to create a repository with sqlmock
func newMockFlightRepo(t *testing.T) (*FlightRepository, sqlmock.Sqlmock, func()) {
	t.Helper()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}

	wrapped := &database.DB{DB: db}

	cleanup := func() {
		db.Close()
	}

	return NewFlightRepository(wrapped), mock, cleanup
}

func TestFlightRepository_SearchFlights_Success(t *testing.T) {
	repo, mock, cleanup := newMockFlightRepo(t)
	defer cleanup()

	req := &models.FlightSearchRequest{
		Source:      "Delhi",
		Destination: "Mumbai",
		Date:        time.Date(2025, 1, 20, 0, 0, 0, 0, time.UTC),
	}

	rows := sqlmock.NewRows([]string{
		"id", "source", "destination", "timestamp",
		"available_seats", "total_seats", "flight_status",
		"price", "version", "created_at", "updated_at",
	}).AddRow(
		int64(1), "Delhi", "Mumbai", time.Now(),
		150, 180, models.FlightStatusScheduled,
		2500.0, 1, time.Now(), time.Now(),
	)

	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT id, source, destination, timestamp, available_seats, total_seats, 
		       flight_status, price, version, created_at, updated_at
		FROM flights
		WHERE source = $1 
		  AND destination = $2 
		  AND DATE(timestamp) = $3
		  AND available_seats > 0
		  AND flight_status IN ('scheduled', 'on_time')
		ORDER BY timestamp ASC
	`)).
		WithArgs(req.Source, req.Destination, req.Date.Format("2006-01-02")).
		WillReturnRows(rows)

	flights, err := repo.SearchFlights(context.Background(), req)
	if err != nil {
		t.Fatalf("SearchFlights returned error: %v", err)
	}

	if len(flights) != 1 {
		t.Fatalf("expected 1 flight, got %d", len(flights))
	}
}

func TestFlightRepository_GetFlightByID_NotFound(t *testing.T) {
	repo, mock, cleanup := newMockFlightRepo(t)
	defer cleanup()

	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT id, source, destination, timestamp, available_seats, total_seats, 
		       flight_status, price, version, created_at, updated_at
		FROM flights
		WHERE id = $1
	`)).
		WithArgs(int64(1)).
		WillReturnError(sql.ErrNoRows)

	flight, err := repo.GetFlightByID(context.Background(), 1)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}

	if flight != nil {
		t.Fatalf("expected nil flight, got %+v", flight)
	}
}

func TestFlightRepository_UpdateAvailableSeats_Success(t *testing.T) {
	repo, mock, cleanup := newMockFlightRepo(t)
	defer cleanup()

	mock.ExpectExec(regexp.QuoteMeta(`
		UPDATE flights 
		SET available_seats = available_seats - $1, 
		    version = version + 1, 
		    updated_at = $2
		WHERE id = $3 AND version = $4 AND available_seats >= $1
	`)).
		WithArgs(2, sqlmock.AnyArg(), int64(1), 1).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.UpdateAvailableSeats(context.Background(), 1, 2, 1)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestFlightRepository_UpdateAvailableSeats_NoRows(t *testing.T) {
	repo, mock, cleanup := newMockFlightRepo(t)
	defer cleanup()

	mock.ExpectExec(regexp.QuoteMeta(`
		UPDATE flights 
		SET available_seats = available_seats - $1, 
		    version = version + 1, 
		    updated_at = $2
		WHERE id = $3 AND version = $4 AND available_seats >= $1
	`)).
		WithArgs(2, sqlmock.AnyArg(), int64(1), 1).
		WillReturnResult(sqlmock.NewResult(0, 0))

	err := repo.UpdateAvailableSeats(context.Background(), 1, 2, 1)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
}

func TestFlightRepository_CreateFlight_Success(t *testing.T) {
	repo, mock, cleanup := newMockFlightRepo(t)
	defer cleanup()

	flight := &models.Flight{
		Source:         "Delhi",
		Destination:    "Mumbai",
		Timestamp:      time.Now(),
		AvailableSeats: 150,
		TotalSeats:     180,
		FlightStatus:   models.FlightStatusScheduled,
		Price:          2500.0,
		Version:        1,
	}

	mock.ExpectQuery(regexp.QuoteMeta(`
		INSERT INTO flights (source, destination, timestamp, available_seats, total_seats, 
		                    flight_status, price, version, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id
	`)).
		WithArgs(
			flight.Source, flight.Destination, flight.Timestamp,
			flight.AvailableSeats, flight.TotalSeats, flight.FlightStatus,
			flight.Price, flight.Version, sqlmock.AnyArg(), sqlmock.AnyArg(),
		).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(int64(1)))

	created, err := repo.CreateFlight(context.Background(), flight)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if created.ID != 1 {
		t.Fatalf("expected id 1, got %d", created.ID)
	}
}

func TestFlightRepository_UpdateFlight_Success(t *testing.T) {
	repo, mock, cleanup := newMockFlightRepo(t)
	defer cleanup()

	flight := &models.Flight{
		ID:             1,
		Source:         "Delhi",
		Destination:    "Mumbai",
		Timestamp:      time.Now(),
		AvailableSeats: 150,
		TotalSeats:     180,
		FlightStatus:   models.FlightStatusScheduled,
		Price:          2500.0,
		Version:        1,
	}

	mock.ExpectExec(regexp.QuoteMeta(`
		UPDATE flights 
		SET source = $1, destination = $2, timestamp = $3, available_seats = $4, 
		    total_seats = $5, flight_status = $6, price = $7, version = version + 1, 
		    updated_at = $8
		WHERE id = $9 AND version = $10
	`)).
		WithArgs(
			flight.Source, flight.Destination, flight.Timestamp, flight.AvailableSeats,
			flight.TotalSeats, flight.FlightStatus, flight.Price, sqlmock.AnyArg(),
			flight.ID, flight.Version,
		).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.UpdateFlight(context.Background(), flight)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestFlightRepository_UpdateFlight_NoRows(t *testing.T) {
	repo, mock, cleanup := newMockFlightRepo(t)
	defer cleanup()

	flight := &models.Flight{
		ID:      1,
		Version: 1,
	}

	mock.ExpectExec(regexp.QuoteMeta(`
		UPDATE flights 
		SET source = $1, destination = $2, timestamp = $3, available_seats = $4, 
		    total_seats = $5, flight_status = $6, price = $7, version = version + 1, 
		    updated_at = $8
		WHERE id = $9 AND version = $10
	`)).
		WithArgs(
			flight.Source, flight.Destination, flight.Timestamp, flight.AvailableSeats,
			flight.TotalSeats, flight.FlightStatus, flight.Price, sqlmock.AnyArg(),
			flight.ID, flight.Version,
		).
		WillReturnResult(sqlmock.NewResult(0, 0))

	err := repo.UpdateFlight(context.Background(), flight)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
}


