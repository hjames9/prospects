CREATE TABLE prospects
(
    id SERIAL8 NOT NULL PRIMARY KEY,
    app_name VARCHAR NOT NULL,
    email VARCHAR NOT NULL,
    first_name VARCHAR NULL,
    last_name VARCHAR NULL,
    phone_number VARCHAR NULL,
    is_valid BOOLEAN NOT NULL DEFAULT FALSE,
    was_processed BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMP NOT NULL
);

CREATE VIEW sneezers
AS
SELECT MAX(id) AS id, app_name, email, MAX(first_name) AS first_name, MAX(last_name) AS last_name, MAX(phone_number) AS phone_number, MAX(created_at) AS created_at FROM prospects WHERE is_valid = TRUE and was_processed = TRUE GROUP BY app_name, email
