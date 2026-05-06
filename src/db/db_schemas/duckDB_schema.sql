CREATE TABLE IF NOT EXISTS raw_sales (
    vin VARCHAR(17) PRIMARY KEY,
    year INTEGER NOT NULL,
    make VARCHAR NOT NULL,
    model VARCHAR NOT NULL,
    trim VARCHAR NOT NULL,
    body VARCHAR NOT NULL,
    transmission VARCHAR NOT NULL,
    color VARCHAR NOT NULL,
    interior VARCHAR NOT NULL,
    selling_price INTEGER NOT NULL,
    mmr INTEGER NOT NULL,
    seller VARCHAR NOT NULL,
    odometer INTEGER NOT NULL,
    condition NUMERIC(3, 2) NOT NULL CHECK (condition >= 0 AND condition <= 5)
);