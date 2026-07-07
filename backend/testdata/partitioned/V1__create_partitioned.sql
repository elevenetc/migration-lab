CREATE TABLE events (
    id SERIAL,
    year INT NOT NULL
) PARTITION BY LIST (year);
