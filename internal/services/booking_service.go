package services

import (
	"context"
	"crypto/rand"
	"fmt"
	"log"
	"time"

	"airline-booking-system/internal/cache"
	"airline-booking-system/internal/config"
	"airline-booking-system/internal/models"
	"airline-booking-system/internal/repositories"
	"airline-booking-system/pkg/kafka"

	"go.opentelemetry.io/otel"
)

// BookingRepository defines persistence operations used by BookingService.
type BookingRepository interface {
	CreateBooking(ctx context.Context, booking *models.Booking) (*models.Booking, error)
	GetBookingByID(ctx context.Context, id int64) (*models.Booking, error)
	GetBookingsByUserID(ctx context.Context, userID int64) ([]models.Booking, error)
	UpdateBookingStatus(ctx context.Context, bookingID int64, status models.BookingStatus, paymentRefID *string) error
}

// FlightRepositoryBooking defines flight operations used by BookingService.
type FlightRepositoryBooking interface {
	GetFlightByID(ctx context.Context, id int64) (*models.Flight, error)
	UpdateAvailableSeats(ctx context.Context, flightID int64, seatsToBook int, version int) error
}

// FlightCacheBooking defines cache operations used by BookingService.
type FlightCacheBooking interface {
	AcquireFlightLock(ctx context.Context, key string) (bool, error)
	ReleaseFlightLock(ctx context.Context, key string) error
	DeleteCachedSeats(ctx context.Context, flightID int64) error
}

// Producer defines the Kafka producer operations used by BookingService.
type Producer interface {
	SendSeatUpdateEvent(ctx context.Context, event *models.SeatUpdateEvent) error
	SendPaymentEvent(ctx context.Context, event *models.PaymentEvent) error
}

// BookingService handles booking business logic
type BookingService struct {
	bookingRepo   BookingRepository
	flightRepo    FlightRepositoryBooking
	cacheService  FlightCacheBooking
	kafkaProducer Producer
	config        *config.AppConfig
	tracerName    string
}

// NewBookingService creates a new booking service
func NewBookingService(
	bookingRepo *repositories.BookingRepository,
	flightRepo *repositories.FlightRepository,
	cacheService *cache.FlightCacheService,
	kafkaProducer *kafka.Producer,
	config *config.AppConfig,
) *BookingService {
	return &BookingService{
		bookingRepo:   bookingRepo,
		flightRepo:    flightRepo,
		cacheService:  cacheService,
		kafkaProducer: kafkaProducer,
		config:        config,
		tracerName:    "airline-booking-system/booking-service",
	}
}

// CreateBooking creates a new booking with distributed locking
func (s *BookingService) CreateBooking(ctx context.Context, req *models.BookingRequest) (*models.BookingResponse, error) {
	tr := otel.Tracer(s.tracerName)
	ctx, span := tr.Start(ctx, "BookingService.CreateBooking")
	defer span.End()

	if !req.IsValid() {
		return nil, fmt.Errorf("invalid booking request")
	}

	// Get flight details
	flight, err := s.flightRepo.GetFlightByID(ctx, req.FlightID)
	if err != nil {
		return nil, fmt.Errorf("failed to get flight: %w", err)
	}

	// Validate flight availability
	if flight.AvailableSeats < req.SeatsBooked {
		return &models.BookingResponse{
			Status:  models.BookingStatusFailed,
			Message: "Insufficient seats available",
		}, nil
	}

	// Check flight status
	if flight.FlightStatus == models.FlightStatusCancelled || flight.FlightStatus == models.FlightStatusDeparted {
		return &models.BookingResponse{
			Status:  models.BookingStatusFailed,
			Message: "Flight is not available for booking",
		}, nil
	}

	// Acquire distributed lock
	lockKey := req.GetLockKey()
	locked, err := s.cacheService.AcquireFlightLock(ctx, lockKey)
	if err != nil {
		return nil, fmt.Errorf("failed to acquire lock: %w", err)
	}

	if !locked {
		return &models.BookingResponse{
			Status:  models.BookingStatusFailed,
			Message: "Flight is currently being booked by another user",
		}, nil
	}

	// Ensure lock is released
	defer func() {
		if err := s.cacheService.ReleaseFlightLock(ctx, lockKey); err != nil {
			log.Printf("Failed to release lock: %v", err)
		}
	}()

	// Double-check seat availability after acquiring lock
	flight, err = s.flightRepo.GetFlightByID(ctx, req.FlightID)
	if err != nil {
		return nil, fmt.Errorf("failed to get flight after lock: %w", err)
	}

	if flight.AvailableSeats < req.SeatsBooked {
		return &models.BookingResponse{
			Status:  models.BookingStatusFailed,
			Message: "Seats no longer available",
		}, nil
	}

	// Calculate booking price
	bookingPrice := flight.Price * float64(req.SeatsBooked)

	// Create booking record with PENDING status
	booking := &models.Booking{
		FlightID:        req.FlightID,
		UserID:          req.UserID,
		Status:          models.BookingStatusPending,
		BookingPrice:    bookingPrice,
		SeatsBooked:     req.SeatsBooked,
		BookingMetadata: req.PassengerDetails,
	}

	createdBooking, err := s.bookingRepo.CreateBooking(ctx, booking)
	if err != nil {
		return nil, fmt.Errorf("failed to create booking: %w", err)
	}

	// Update available seats in database
	err = s.flightRepo.UpdateAvailableSeats(ctx, req.FlightID, req.SeatsBooked, flight.Version)
	if err != nil {
		// If seat update fails, mark booking as failed
		s.bookingRepo.UpdateBookingStatus(ctx, createdBooking.ID, models.BookingStatusFailed, nil)
		return &models.BookingResponse{
			BookingID: createdBooking.ID,
			Status:    models.BookingStatusFailed,
			Message:   "Failed to reserve seats",
		}, nil
	}

	// Invalidate cache for this flight's seats
	s.cacheService.DeleteCachedSeats(ctx, req.FlightID)

	// Generate payment reference ID
	paymentRefID := generatePaymentReferenceID()

	// Simulate payment processing (in real implementation, this would call payment gateway)
	go s.processPaymentAsync(ctx, createdBooking.ID, paymentRefID, bookingPrice)

	return &models.BookingResponse{
		BookingID:         createdBooking.ID,
		Status:           models.BookingStatusPending,
		PaymentReferenceID: paymentRefID,
		Message:          "Booking created, processing payment",
	}, nil
}

// processPaymentAsync simulates async payment processing
func (s *BookingService) processPaymentAsync(ctx context.Context, bookingID int64, paymentRefID string, amount float64) {
	tr := otel.Tracer(s.tracerName)
	ctx, span := tr.Start(ctx, "BookingService.processPaymentAsync")
	defer span.End()

	// Simulate payment processing delay
	time.Sleep(2 * time.Second)

	// Simulate payment success (90% success rate)
	paymentSuccessful := simulatePaymentSuccess()

	var newStatus models.BookingStatus
	var message string

	if paymentSuccessful {
		newStatus = models.BookingStatusCompleted
		message = "Payment successful"

		// Send seat update event
		seatEvent := &models.SeatUpdateEvent{
			FlightID:    0, // Would be retrieved from booking
			SeatsBooked: 0, // Would be retrieved from booking
			Timestamp:   time.Now(),
			BookingID:   bookingID,
		}

		if err := s.kafkaProducer.SendSeatUpdateEvent(ctx, seatEvent); err != nil {
			log.Printf("Failed to send seat update event: %v", err)
		}
	} else {
		newStatus = models.BookingStatusFailed
		message = "Payment failed"

		// In a real implementation, you would need to release the seats back
		// This is simplified for the demo
	}

	// Update booking status
	err := s.bookingRepo.UpdateBookingStatus(ctx, bookingID, newStatus, &paymentRefID)
	if err != nil {
		log.Printf("Failed to update booking status: %v", err)
		return
	}

	// Send payment event
	paymentEvent := &models.PaymentEvent{
		BookingID:         bookingID,
		PaymentReferenceID: paymentRefID,
		Amount:           amount,
		Status:           string(newStatus),
		Timestamp:        time.Now(),
	}

	if err := s.kafkaProducer.SendPaymentEvent(ctx, paymentEvent); err != nil {
		log.Printf("Failed to send payment event: %v", err)
	}

	log.Printf("Booking %d payment processing completed: %s", bookingID, message)
}

// GetBookingByID gets a booking by ID
func (s *BookingService) GetBookingByID(ctx context.Context, id int64) (*models.Booking, error) {
	return s.bookingRepo.GetBookingByID(ctx, id)
}

// GetBookingsByUserID gets bookings for a user
func (s *BookingService) GetBookingsByUserID(ctx context.Context, userID int64) ([]models.Booking, error) {
	return s.bookingRepo.GetBookingsByUserID(ctx, userID)
}

// generatePaymentReferenceID generates a unique payment reference ID
func generatePaymentReferenceID() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return fmt.Sprintf("PAY-%x", bytes)
}

// simulatePaymentSuccess simulates payment success/failure (90% success rate)
func simulatePaymentSuccess() bool {
	return time.Now().UnixNano()%10 != 0
}
