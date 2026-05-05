
CREATE TABLE dim_year (
    year_id INTEGER PRIMARY KEY,
    year INTEGER NOT NULL
);

CREATE TABLE dim_make (
    make_id INTEGER PRIMARY KEY,
    make_name VARCHAR NOT NULL UNIQUE
);

CREATE TABLE dim_model (
    model_id INTEGER PRIMARY KEY,
    model_name VARCHAR NOT NULL UNIQUE
);

CREATE TABLE dim_trim (
    trim_id INTEGER PRIMARY KEY,
    trim_name VARCHAR NOT NULL UNIQUE
);

CREATE TABLE dim_body (
    body_id INTEGER PRIMARY KEY,
    body_name VARCHAR NOT NULL UNIQUE
);

CREATE TABLE dim_transmission (
    transmission_id INTEGER PRIMARY KEY,
    transmission_name VARCHAR NOT NULL UNIQUE
);

CREATE TABLE dim_color (
    color_id INTEGER PRIMARY KEY,
    color_name VARCHAR NOT NULL UNIQUE
);

CREATE TABLE dim_interior (
    interior_id INTEGER PRIMARY KEY,
    interior_name VARCHAR NOT NULL UNIQUE
);

CREATE TABLE fact_sales (
    vin VARCHAR(17) PRIMARY KEY,
    year_id INTEGER NOT NULL,
    FOREIGN KEY (year_id) REFERENCES dim_year(year_id),
    make_id INTEGER NOT NULL,
    FOREIGN KEY (make_id) REFERENCES dim_make(make_id),
    model_id INTEGER NOT NULL,
    FOREIGN KEY (model_id) REFERENCES dim_model(model_id),
    trim_id INTEGER NOT NULL,
    FOREIGN KEY (trim_id) REFERENCES dim_trim(trim_id),
    body_id INTEGER NOT NULL,
    FOREIGN KEY (body_id) REFERENCES dim_body(body_id),
    transmission_id INTEGER NOT NULL,
    FOREIGN KEY (transmission_id) REFERENCES dim_transmission(transmission_id),
    color_id INTEGER NOT NULL,
    FOREIGN KEY (color_id) REFERENCES dim_color(color_id),
    interior_id INTEGER NOT NULL,
    FOREIGN KEY (interior_id) REFERENCES dim_interior(interior_id),
    selling_price INTEGER NOT NULL,
    mmr INTEGER NOT NULL,
    seller VARCHAR(100) NOT NULL,
    odometer INTEGER NOT NULL,
    condition NUMERIC(3, 2) NOT NULL CHECK (condition >= 0 AND condition <= 5)
);