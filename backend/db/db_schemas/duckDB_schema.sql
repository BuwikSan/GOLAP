CREATE SEQUENCE id_sequence START 1;

CREATE TABLE IF NOT EXISTS raw_sales (
    id INTEGER NOT NULL DEFAULT nextval('id_sequence') PRIMARY KEY,
    vin VARCHAR(17),
    year_int INTEGER,
    make VARCHAR,
    model VARCHAR,
    trim VARCHAR,
    body VARCHAR,
    transmission VARCHAR,
    color VARCHAR,
    interior VARCHAR,
    selling_price INTEGER,
    mmr INTEGER,
    seller VARCHAR,
    odometer INTEGER,
    condition NUMERIC(3, 2) CHECK (
        (
            condition >= 0
            AND condition <= 5
        )
        OR (NULL)
    )
);