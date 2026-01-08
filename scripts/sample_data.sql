-- Insert sample flights for testing
INSERT INTO flights (source, destination, timestamp, available_seats, total_seats, flight_status, price) VALUES
('Delhi', 'Mumbai', '2025-01-15 10:00:00', 150, 180, 'scheduled', 2500.00),
('Delhi', 'Mumbai', '2025-01-15 14:00:00', 120, 180, 'scheduled', 2500.00),
('Delhi', 'Mumbai', '2025-01-15 18:00:00', 180, 180, 'scheduled', 2500.00),
('Delhi', 'Bangalore', '2025-01-15 08:00:00', 100, 150, 'scheduled', 3200.00),
('Delhi', 'Bangalore', '2025-01-15 12:00:00', 80, 150, 'scheduled', 3200.00),
('Delhi', 'Chennai', '2025-01-15 09:00:00', 90, 120, 'scheduled', 2800.00),
('Delhi', 'Chennai', '2025-01-15 16:00:00', 120, 120, 'scheduled', 2800.00),
('Mumbai', 'Delhi', '2025-01-15 11:00:00', 140, 180, 'scheduled', 2500.00),
('Mumbai', 'Delhi', '2025-01-15 15:00:00', 160, 180, 'scheduled', 2500.00),
('Mumbai', 'Bangalore', '2025-01-15 07:00:00', 110, 150, 'scheduled', 2200.00),
('Bangalore', 'Delhi', '2025-01-15 13:00:00', 95, 150, 'scheduled', 3200.00),
('Chennai', 'Delhi', '2025-01-15 17:00:00', 85, 120, 'scheduled', 2800.00),
-- Future dates for testing
('Delhi', 'Mumbai', '2025-01-20 10:00:00', 180, 180, 'scheduled', 2500.00),
('Delhi', 'Bangalore', '2025-01-20 08:00:00', 150, 150, 'scheduled', 3200.00),
('Mumbai', 'Chennai', '2025-01-20 09:00:00', 120, 120, 'scheduled', 2400.00);
