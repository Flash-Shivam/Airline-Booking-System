-- Create flights table
CREATE TABLE IF NOT EXISTS flights (
    id BIGSERIAL PRIMARY KEY,
    source VARCHAR(100) NOT NULL,
    destination VARCHAR(100) NOT NULL,
    timestamp TIMESTAMP NOT NULL,
    available_seats INTEGER NOT NULL CHECK (available_seats >= 0),
    total_seats INTEGER NOT NULL CHECK (total_seats > 0),
    flight_status VARCHAR(50) NOT NULL DEFAULT 'scheduled',
    price DECIMAL(10,2) NOT NULL CHECK (price > 0),
    version INTEGER NOT NULL DEFAULT 1,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    CONSTRAINT chk_available_seats CHECK (available_seats <= total_seats),
    CONSTRAINT chk_source_destination CHECK (source != destination)
);

-- Create index for flight search
CREATE INDEX IF NOT EXISTS idx_flights_search ON flights(source, destination, DATE(timestamp));
CREATE INDEX IF NOT EXISTS idx_flights_timestamp ON flights(timestamp);
CREATE INDEX IF NOT EXISTS idx_flights_status ON flights(flight_status);

-- Create bookings table
CREATE TABLE IF NOT EXISTS bookings (
    id BIGSERIAL PRIMARY KEY,
    flight_id BIGINT NOT NULL REFERENCES flights(id),
    user_id BIGINT NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    payment_reference_id VARCHAR(255),
    booking_price DECIMAL(10,2) NOT NULL CHECK (booking_price > 0),
    seats_booked INTEGER NOT NULL CHECK (seats_booked > 0),
    booking_metadata JSONB,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for bookings
CREATE INDEX IF NOT EXISTS idx_bookings_flight_id ON bookings(flight_id);
CREATE INDEX IF NOT EXISTS idx_bookings_user_id ON bookings(user_id);
CREATE INDEX IF NOT EXISTS idx_bookings_status ON bookings(status);
CREATE INDEX IF NOT EXISTS idx_bookings_created_at ON bookings(created_at DESC);

-- Create function to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Create triggers for updated_at
CREATE TRIGGER update_flights_updated_at BEFORE UPDATE ON flights 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_bookings_updated_at BEFORE UPDATE ON bookings 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
