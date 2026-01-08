package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"airline-booking-system/internal/cache"
	"airline-booking-system/internal/config"
	"airline-booking-system/internal/handlers"
	"airline-booking-system/internal/repositories"
	"airline-booking-system/internal/services"
	"airline-booking-system/pkg/database"
	"airline-booking-system/pkg/kafka"
	"airline-booking-system/pkg/redis"

	"github.com/gorilla/mux"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Initialize database
	db, err := database.NewPostgresConnection(&cfg.Database)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Initialize Redis
	redisClient := redis.NewClient(&cfg.Redis)
	defer redisClient.Close()

	// Test Redis connection
	if err := redisClient.Ping(context.Background()); err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}

	// Initialize Kafka producer
	kafkaProducer := kafka.NewProducer(&cfg.Kafka)
	defer kafkaProducer.Close()

	// Initialize repositories
	flightRepo := repositories.NewFlightRepository(db)
	bookingRepo := repositories.NewBookingRepository(db)

	// Initialize cache service
	cacheService := cache.NewFlightCacheService(redisClient, &cfg.App)

	// Initialize services
	flightService := services.NewFlightService(flightRepo, cacheService, &cfg.App)
	bookingService := services.NewBookingService(bookingRepo, flightRepo, cacheService, kafkaProducer, &cfg.App)

	// Initialize handlers
	flightHandler := handlers.NewFlightHandler(flightService)
	bookingHandler := handlers.NewBookingHandler(bookingService)

	// Setup routes
	router := setupRoutes(flightHandler, bookingHandler)

	// Setup server
	server := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      router,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
	}

	// Start server in a goroutine
	go func() {
		log.Printf("Starting server on port %s", cfg.Server.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	// Give outstanding requests 30 seconds to complete
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}

func setupRoutes(fh *handlers.FlightHandler, bh *handlers.BookingHandler) *mux.Router {
	router := mux.NewRouter()

	// API version prefix
	api := router.PathPrefix("/api/v1").Subrouter()

	// Flight routes
	api.HandleFunc("/flights/search", fh.SearchFlights).Methods("GET")
	api.HandleFunc("/flights/{id}", fh.GetFlight).Methods("GET")
	api.HandleFunc("/flights", fh.CreateFlight).Methods("POST")
	api.HandleFunc("/flights/{id}", fh.UpdateFlight).Methods("PUT")

	// Booking routes
	api.HandleFunc("/bookings", bh.CreateBooking).Methods("POST")
	api.HandleFunc("/bookings/{id}", bh.GetBooking).Methods("GET")
	api.HandleFunc("/users/{userId}/bookings", bh.GetUserBookings).Methods("GET")

	// Health check
	api.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}).Methods("GET")

	// Add middleware
	router.Use(loggingMiddleware)
	router.Use(corsMiddleware)

	return router
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s %v", r.Method, r.URL.Path, time.Since(start))
	})
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}
