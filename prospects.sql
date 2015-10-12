DROP SCHEMA IF EXISTS prospects CASCADE;

CREATE SCHEMA IF NOT EXISTS prospects;

SET search_path TO prospects,public;

CREATE TYPE gender AS ENUM ('male', 'female');

CREATE TABLE leads
(
    id SERIAL8 NOT NULL PRIMARY KEY,
    lead_id UUID NOT NULL,
    app_name VARCHAR NOT NULL,
    email VARCHAR NULL,
    used_pinterest BOOLEAN NOT NULL DEFAULT FALSE,
    used_facebook BOOLEAN NOT NULL DEFAULT FALSE,
    used_instagram BOOLEAN NOT NULL DEFAULT FALSE,
    used_twitter BOOLEAN NOT NULL DEFAULT FALSE,
    used_google BOOLEAN NOT NULL DEFAULT FALSE,
    used_youtube BOOLEAN NOT NULL DEFAULT FALSE,
    feedback VARCHAR NULL,
    referrer VARCHAR NULL,
    page_referrer VARCHAR NULL,
    first_name VARCHAR NULL,
    last_name VARCHAR NULL,
    phone_number VARCHAR NULL,
    dob DATE NULL,
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
    created_at TIMESTAMP NOT NULL,
    CHECK(email ~* '^[A-Za-z0-9._%-]+@[A-Za-z0-9.-]+[.][A-Za-z]+$'),
    CHECK(geolocation[0] >= -90.0 AND geolocation[0] <= 90.0 AND geolocation[1] >= -180.0 AND geolocation[1] <= 180.0),
    CHECK(email IS NOT NULL OR used_pinterest IS TRUE OR used_facebook IS TRUE OR used_instagram IS TRUE OR used_twitter IS TRUE OR used_google IS TRUE OR used_youtube IS TRUE OR feedback IS NOT NULL)
);

ALTER SEQUENCE leads_id_seq INCREMENT BY 7 START WITH 31337 RESTART WITH 31337;

CREATE VIEW sneezers
AS
SELECT MAX(id) AS id,
       lead_id,
       app_name,
       MAX(email) AS email,
       BOOL_OR(used_pinterest) AS used_pinterest,
       BOOL_OR(used_facebook) AS used_facebook,
       BOOL_OR(used_instagram) AS used_instagram,
       BOOL_OR(used_twitter) AS used_twitter,
       BOOL_OR(used_google) AS used_google,
       BOOL_OR(used_youtube) AS used_youtube,
       MAX(feedback) AS feedback,
       MAX(first_name) AS first_name,
       MAX(last_name) AS last_name,
       MAX(phone_number) AS phone_number,
       MAX(dob) AS dob,
       MAX(gender) AS gender,
       MAX(zip_code) AS zip_code,
       MAX(language) AS language,
       MAX(user_agent) AS user_agent,
       MAX(created_at) AS created_at
FROM leads
WHERE is_valid = TRUE AND was_processed = TRUE
GROUP BY lead_id, app_name, email;

CREATE INDEX l_lead_id_idx ON leads(lead_id);

CREATE INDEX l_app_name_idx ON leads(app_name);

CREATE INDEX l_email_idx ON leads(email);

CREATE INDEX l_referrer_idx ON leads(page_referrer);

CREATE INDEX l_misc_idx ON leads USING GIN(miscellaneous);
