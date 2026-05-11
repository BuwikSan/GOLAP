CREATE TABLE dim_year (
    year_id SERIAL PRIMARY KEY,
    year_int INTEGER UNIQUE
);

CREATE TABLE dim_make (
    make_id SERIAL PRIMARY KEY,
    make_name VARCHAR UNIQUE
);

CREATE TABLE dim_model (
    model_id SERIAL PRIMARY KEY,
    model_name VARCHAR UNIQUE
);

CREATE TABLE dim_trim (
    trim_id SERIAL PRIMARY KEY,
    trim_name VARCHAR UNIQUE
);

CREATE TABLE dim_body (
    body_id SERIAL PRIMARY KEY,
    body_name VARCHAR UNIQUE
);

CREATE TABLE dim_transmission (
    transmission_id SERIAL PRIMARY KEY,
    transmission_name VARCHAR UNIQUE
);

CREATE TABLE dim_color (
    color_id SERIAL PRIMARY KEY,
    color_name VARCHAR UNIQUE
);

CREATE TABLE dim_interior (
    interior_id SERIAL PRIMARY KEY,
    interior_name VARCHAR UNIQUE
);

CREATE TABLE fact_sales (
    id SERIAL PRIMARY KEY,
    vin VARCHAR(17),
    year_id INTEGER NOT NULL,
    FOREIGN KEY (year_id) REFERENCES dim_year (year_id),
    make_id INTEGER NOT NULL,
    FOREIGN KEY (make_id) REFERENCES dim_make (make_id),
    model_id INTEGER NOT NULL,
    FOREIGN KEY (model_id) REFERENCES dim_model (model_id),
    trim_id INTEGER NOT NULL,
    FOREIGN KEY (trim_id) REFERENCES dim_trim (trim_id),
    body_id INTEGER NOT NULL,
    FOREIGN KEY (body_id) REFERENCES dim_body (body_id),
    transmission_id INTEGER NOT NULL,
    FOREIGN KEY (transmission_id) REFERENCES dim_transmission (transmission_id),
    color_id INTEGER NOT NULL,
    FOREIGN KEY (color_id) REFERENCES dim_color (color_id),
    interior_id INTEGER NOT NULL,
    FOREIGN KEY (interior_id) REFERENCES dim_interior (interior_id),
    selling_price INTEGER NOT NULL,
    mmr INTEGER NOT NULL,
    seller VARCHAR(100),
    odometer INTEGER,
    condition NUMERIC(3, 2) CHECK (
        (
            condition >= 0
            AND condition <= 5
        )
        OR (NULL)
    )
);