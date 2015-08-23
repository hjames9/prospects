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
SELECT id, app_name, email, first_name, last_name, phone_number, created_at FROM prospects WHERE is_valid = TRUE and was_processed = TRUE
