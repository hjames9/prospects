CREATE TYPE gender AS ENUM ('male', 'female');

CREATE TABLE prospects
(
    id SERIAL8 NOT NULL PRIMARY KEY,
    app_name VARCHAR NOT NULL,
    email VARCHAR NOT NULL,
    referrer VARCHAR NULL,
    page_referrer VARCHAR NULL,
    first_name VARCHAR NULL,
    last_name VARCHAR NULL,
    phone_number VARCHAR NULL,
    age SMALLINT NULL,
    gender GENDER NULL,
    zip_code VARCHAR NULL,
    language VARCHAR NULL,
    user_agent VARCHAR NULL,
    cookies VARCHAR[] NULL,
    geolocation POINT NULL,
    ip_address INET NULL,
    miscellaneous JSONB NULL,
    is_valid BOOLEAN NOT NULL DEFAULT FALSE,
    was_processed BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMP NOT NULL
);

CREATE VIEW sneezers
AS
SELECT MAX(id) AS id, app_name, email, MAX(first_name) AS first_name, MAX(last_name) AS last_name, MAX(phone_number) AS phone_number, MAX(age) AS age, MAX(gender) AS gender, MAX(zip_code) AS zip_code, MAX(user_agent) AS user_agent, MAX(created_at) AS created_at FROM prospects WHERE is_valid = TRUE AND was_processed = TRUE GROUP BY app_name, email;

CREATE INDEX p_app_name_idx ON prospects(app_name);

CREATE INDEX p_email_idx ON prospects(email);

CREATE INDEX p_referrer_idx ON prospects(page_referrer);

CREATE INDEX p_misc_idx ON prospects USING GIN(miscellaneous);
