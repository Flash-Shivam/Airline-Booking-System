package services

import (
	"context"
	"testing"

	"airline-booking-system/internal/models"
)

// mockBookingRepo implements BookingRepository for testing.
type mockBookingRepo struct {
	createFn           func(ctx context.Context, booking *models.Booking) (*models.Booking, error)
	getByIDFn          func(ctx context.Context, id int64) (*models.Booking, error)
	getByUserFn        func(ctx context.Context, userID int64) ([]models.Booking, error)
	updateStatusFn     func(ctx context.Context, bookingID int64, status models.BookingStatus, paymentRefID *string) error
}

func (m *mockBookingRepo) CreateBooking(ctx context.Context, booking *models.Booking) (*models.Booking, error) {
	if m.createFn != nil {
		return m.createFn(ctx, booking)
	}
	return booking, nil
}

func (m *mockBookingRepo) GetBookingByID(ctx context.Context, id int64) (*models.Booking, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(ctx, id)
	}
	return nil, nil
}

func (m *mockBookingRepo) GetBookingsByUserID(ctx context.Context, userID int64) ([]models.Booking, error) {
	if m.getByUserFn != nil {
		return m.getByUserFn(ctx, userID)
	}
	return nil, nil
}

func (m *mockBookingRepo) UpdateBookingStatus(ctx context.Context, bookingID int64, status models.BookingStatus, paymentRefID *string) error {
	if m.updateStatusFn != nil {
		return m.updateStatusFn(ctx, bookingID, status, paymentRefID)
	}
	return nil
}

// mockFlightRepoBooking implements FlightRepositoryBooking for testing.
type mockFlightRepoBooking struct {
	getByIDFn           func(ctx context.Context, id int64) (*models.Flight, error)
	updateAvailableFn   func(ctx context.Context, flightID int64, seatsToBook int, version int) error
}

func (m *mockFlightRepoBooking) GetFlightByID(ctx context.Context, id int64) (*models.Flight, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(ctx, id)
	}
	return nil, nil
}

func (m *mockFlightRepoBooking) UpdateAvailableSeats(ctx context.Context, flightID int64, seatsToBook int, version int) error {
	if m.updateAvailableFn != nil {
		return m.updateAvailableFn(ctx, flightID, seatsToBook, version)
	}
	return nil
}

// mockFlightCacheBooking implements FlightCacheBooking for testing.
type mockFlightCacheBooking struct {
	acquireFn func(ctx context.Context, key string) (bool, error)
	releaseFn func(ctx context.Context, key string) error
	deleteFn  func(ctx context.Context, flightID int64) error
}

func (m *mockFlightCacheBooking) AcquireFlightLock(ctx context.Context, key string) (bool, error) {
	if m.acquireFn != nil {
		return m.acquireFn(ctx, key)
	}
	return true, nil
}

func (m *mockFlightCacheBooking) ReleaseFlightLock(ctx context.Context, key string) error {
	if m.releaseFn != nil {
		return m.releaseFn(ctx, key)
	}
	return nil
}

func (m *mockFlightCacheBooking) DeleteCachedSeats(ctx context.Context, flightID int64) error {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, flightID)
	}
	return nil
}

// mockProducer implements Producer for testing.
type mockProducer struct {
	sendSeatFn    func(ctx context.Context, event *models.SeatUpdateEvent) error
	sendPaymentFn func(ctx context.Context, event *models.PaymentEvent) error
}

func (m *mockProducer) SendSeatUpdateEvent(ctx context.Context, event *models.SeatUpdateEvent) error {
	if m.sendSeatFn != nil {
		return m.sendSeatFn(ctx, event)
	}
	return nil
}

func (m *mockProducer) SendPaymentEvent(ctx context.Context, event *models.PaymentEvent) error {
	if m.sendPaymentFn != nil {
		return m.sendPaymentFn(ctx, event)
	}
	return nil
}

func TestBookingService_CreateBooking_InvalidRequest(t *testing.T) {
	svc := &BookingService{}

	req := &models.BookingRequest{} // invalid: missing fields
	if _, err := svc.CreateBooking(context.Background(), req); err == nil {
		t.Fatalf("expected error for invalid request, got nil")
	}
}

func TestBookingService_CreateBooking_InsufficientSeats(t *testing.T) {
	bookingRepo := &mockBookingRepo{}
	flightRepo := &mockFlightRepoBooking{
		getByIDFn: func(ctx context.Context, id int64) (*models.Flight, error) {
			return &models.Flight{
				ID:             id,
				AvailableSeats: 1,
				TotalSeats:     10,
				Price:          100,
			}, nil
		},
	}
	cache := &mockFlightCacheBooking{}
	producer := &mockProducer{}

	svc := &BookingService{
		bookingRepo:   bookingRepo,
		flightRepo:    flightRepo,
		cacheService:  cache,
		kafkaProducer: producer,
	}

	req := &models.BookingRequest{
		FlightID: 1,
		UserID:   123,
		SeatsBooked: 2,
		PassengerDetails: []models.PassengerDetails{
			{Name: "John"},
			{Name: "Jane"},
		},
	}

	resp, err := svc.CreateBooking(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp.Status != models.BookingStatusFailed {
		t.Fatalf("expected failed status, got %s", resp.Status)
	}
}

func TestBookingService_CreateBooking_FlightNotBookable(t *testing.T) {
	bookingRepo := &mockBookingRepo{}
	flightRepo := &mockFlightRepoBooking{
		getByIDFn: func(ctx context.Context, id int64) (*models.Flight, error) {
			return &models.Flight{
				ID:             id,
				AvailableSeats: 10,
				TotalSeats:     10,
				Price:          100,
				FlightStatus:   models.FlightStatusCancelled,
			}, nil
		},
	}
	cache := &mockFlightCacheBooking{}
	producer := &mockProducer{}

	svc := &BookingService{
		bookingRepo:   bookingRepo,
		flightRepo:    flightRepo,
		cacheService:  cache,
		kafkaProducer: producer,
	}

	req := &models.BookingRequest{
		FlightID: 1,
		UserID:   123,
		SeatsBooked: 1,
		PassengerDetails: []models.PassengerDetails{
			{Name: "John"},
		},
	}

	resp, err := svc.CreateBooking(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp.Status != models.BookingStatusFailed {
		t.Fatalf("expected failed status, got %s", resp.Status)
	}
}

func TestBookingService_CreateBooking_LockNotAcquired(t *testing.T) {
	bookingRepo := &mockBookingRepo{}
	flightRepo := &mockFlightRepoBooking{
		getByIDFn: func(ctx context.Context, id int64) (*models.Flight, error) {
			return &models.Flight{
				ID:             id,
				AvailableSeats: 10,
				TotalSeats:     10,
				Price:          100,
				FlightStatus:   models.FlightStatusScheduled,
			}, nil
		},
	}
	cache := &mockFlightCacheBooking{
		acquireFn: func(ctx context.Context, key string) (bool, error) {
			return false, nil
		},
	}
	producer := &mockProducer{}

	svc := &BookingService{
		bookingRepo:   bookingRepo,
		flightRepo:    flightRepo,
		cacheService:  cache,
		kafkaProducer: producer,
	}

	req := &models.BookingRequest{
		FlightID: 1,
		UserID:   123,
		SeatsBooked: 1,
		PassengerDetails: []models.PassengerDetails{
			{Name: "John"},
		},
	}

	resp, err := svc.CreateBooking(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp.Status != models.BookingStatusFailed {
		t.Fatalf("expected failed status, got %s", resp.Status)
	}
}

func TestBookingService_CreateBooking_FlightSeatsGoneAfterLock(t *testing.T) {
	call := 0
	bookingRepo := &mockBookingRepo{}
	flightRepo := &mockFlightRepoBooking{
		getByIDFn: func(ctx context.Context, id int64) (*models.Flight, error) {
			call++
			if call == 1 {
				return &models.Flight{
					ID:             id,
					AvailableSeats: 10,
					TotalSeats:     10,
					Price:          100,
					FlightStatus:   models.FlightStatusScheduled,
				}, nil
			}
			return &models.Flight{
				ID:             id,
				AvailableSeats: 0,
				TotalSeats:     10,
				Price:          100,
				FlightStatus:   models.FlightStatusScheduled,
			}, nil
		},
	}
	cache := &mockFlightCacheBooking{}
	producer := &mockProducer{}

	svc := &BookingService{
		bookingRepo:   bookingRepo,
		flightRepo:    flightRepo,
		cacheService:  cache,
		kafkaProducer: producer,
	}

	req := &models.BookingRequest{
		FlightID: 1,
		UserID:   123,
		SeatsBooked: 1,
		PassengerDetails: []models.PassengerDetails{
			{Name: "John"},
		},
	}

	resp, err := svc.CreateBooking(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp.Status != models.BookingStatusFailed {
		t.Fatalf("expected failed status, got %s", resp.Status)
	}
}

func TestBookingService_CreateBooking_SuccessBasicFlow(t *testing.T) {
	bookingCreated := false
	seatsUpdated := false
	cacheDeleted := false

	bookingRepo := &mockBookingRepo{
		createFn: func(ctx context.Context, booking *models.Booking) (*models.Booking, error) {
			bookingCreated = true
			booking.ID = 1
			return booking, nil
		},
	}
	flightRepo := &mockFlightRepoBooking{
		getByIDFn: func(ctx context.Context, id int64) (*models.Flight, error) {
			return &models.Flight{
				ID:             id,
				AvailableSeats: 10,
				TotalSeats:     10,
				Price:          100,
				FlightStatus:   models.FlightStatusScheduled,
				Version:        1,
			}, nil
		},
		updateAvailableFn: func(ctx context.Context, flightID int64, seatsToBook int, version int) error {
			seatsUpdated = true
			return nil
		},
	}
	cache := &mockFlightCacheBooking{
		deleteFn: func(ctx context.Context, flightID int64) error {
			cacheDeleted = true
			return nil
		},
	}
	producer := &mockProducer{}

	svc := &BookingService{
		bookingRepo:   bookingRepo,
		flightRepo:    flightRepo,
		cacheService:  cache,
		kafkaProducer: producer,
	}

	req := &models.BookingRequest{
		FlightID: 1,
		UserID:   123,
		SeatsBooked: 2,
		PassengerDetails: []models.PassengerDetails{
			{Name: "John"},
			{Name: "Jane"},
		},
	}

	resp, err := svc.CreateBooking(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp.Status != models.BookingStatusPending {
		t.Fatalf("expected pending status, got %s", resp.Status)
	}

	if !bookingCreated || !seatsUpdated || !cacheDeleted {
		t.Fatalf("expected bookingCreated=%v, seatsUpdated=%v, cacheDeleted=%v to all be true", bookingCreated, seatsUpdated, cacheDeleted)
	}
}

func TestBookingService_GetBookingByID_DelegatesToRepo(t *testing.T) {
	called := false
	bookingRepo := &mockBookingRepo{
		getByIDFn: func(ctx context.Context, id int64) (*models.Booking, error) {
			called = true
			return &models.Booking{ID: id}, nil
		},
	}
	svc := &BookingService{
		bookingRepo: bookingRepo,
	}

	booking, err := svc.GetBookingByID(context.Background(), 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !called || booking.ID != 1 {
		t.Fatalf("expected repo to be called and booking id 1")
	}
}

func TestBookingService_GetBookingsByUserID_DelegatesToRepo(t *testing.T) {
	called := false
	bookingRepo := &mockBookingRepo{
		getByUserFn: func(ctx context.Context, userID int64) ([]models.Booking, error) {
			called = true
			return []models.Booking{{ID: 1}}, nil
		},
	}
	svc := &BookingService{
		bookingRepo: bookingRepo,
	}

	bookings, err := svc.GetBookingsByUserID(context.Background(), 123)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !called || len(bookings) != 1 {
		t.Fatalf("expected repo to be called and one booking returned")
	}
}


