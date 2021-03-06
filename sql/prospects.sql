DROP SCHEMA IF EXISTS prospects CASCADE;

CREATE SCHEMA IF NOT EXISTS prospects;

SET search_path TO prospects,public;

CREATE TYPE gender AS ENUM ('male', 'female');

CREATE TYPE lead_source AS ENUM ('landing', 'email', 'phone', 'extended', 'feedback', 'pinterest', 'facebook', 'instagram', 'twitter', 'google', 'snapchat', 'youtube', 'popup');

CREATE TABLE leads
(
    id SERIAL8 NOT NULL PRIMARY KEY,
    lead_id UUID NOT NULL,
    app_name VARCHAR NOT NULL,
    email VARCHAR NULL,
    lead_source LEAD_SOURCE NOT NULL,
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
    replied_to BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    CHECK(email ~* '^[A-Za-z0-9._%-]+@[A-Za-z0-9.-]+[.][A-Za-z]+$'),
    CHECK(geolocation[0] >= -90.0 AND geolocation[0] <= 90.0 AND geolocation[1] >= -180.0 AND geolocation[1] <= 180.0),
    CHECK(lead_source <> 'landing' OR (lead_source = 'landing' AND (email IS NOT NULL OR phone_number IS NOT NULL))),
    CHECK(lead_source <> 'phone' OR (lead_source = 'phone' AND phone_number IS NOT NULL)),
    CHECK(lead_source <> 'email' OR (lead_source = 'email' AND email IS NOT NULL)),
    CHECK(lead_source <> 'feedback' OR (lead_source = 'feedback' AND feedback IS NOT NULL)),
    CHECK(lead_source <> 'extended' OR (lead_source = 'extended' AND (first_name IS NOT NULL OR last_name IS NOT NULL OR dob IS NOT NULL OR gender IS NOT NULL OR zip_code IS NOT NULL OR language IS NOT NULL OR miscellaneous IS NOT NULL)))
);

ALTER SEQUENCE leads_id_seq INCREMENT BY 7 START WITH 31337 RESTART WITH 31337;

CREATE OR REPLACE VIEW sneezers
AS
SELECT MAX(id) AS id,
       lead_id,
       app_name,
       email,
       replied_to,
       MAX(lead_source) AS lead_source,
       MAX(feedback) AS feedback,
       MAX(first_name) AS first_name,
       MAX(last_name) AS last_name,
       MAX(phone_number) AS phone_number,
       MAX(dob) AS dob,
       MAX(gender) AS gender,
       MAX(zip_code) AS zip_code,
       MAX(language) AS language,
       MAX(user_agent) AS user_agent,
       JSON_AGG(miscellaneous) AS miscellaneous,
       BOOL_OR(was_processed) AS was_processed,
       BOOL_OR(is_valid) AS is_valid,
       MAX(created_at) AS created_at,
       MAX(updated_at) AS updated_at
FROM leads
WHERE is_valid = TRUE AND was_processed = TRUE
GROUP BY lead_id, app_name, email, replied_to;

CREATE INDEX l_lead_id_idx ON leads(lead_id);

CREATE INDEX l_app_name_idx ON leads(app_name);

CREATE INDEX l_email_idx ON leads(email);

CREATE INDEX l_referrer_idx ON leads(page_referrer);

CREATE INDEX l_misc_idx ON leads USING GIN(miscellaneous);

CREATE TABLE imap_markers
(
    app_name VARCHAR NOT NULL PRIMARY KEY,
    marker INT8 NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    CHECK(marker > 0)
);

CREATE TABLE mailer_queries
(
    mailer_name VARCHAR NOT NULL PRIMARY KEY,
    source_email_address VARCHAR NOT NULL,
    get_email_data_query VARCHAR NOT NULL,
    dest_email_field_name VARCHAR NOT NULL,
    email_subject VARCHAR NOT NULL,
    email_subject_field_names VARCHAR[] NULL,
    email_template_url VARCHAR NOT NULL,
    update_status_query VARCHAR NULL,
    update_status_identifer VARCHAR NULL,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    CHECK(source_email_address ~* '^[A-Za-z0-9._%-]+@[A-Za-z0-9.-]+[.][A-Za-z]+$'),
    CHECK(email_subject_field_names IS NOT NULL OR (email_subject_field_names IS NULL AND email_subject !~* '%\S*')),
    CHECK(email_template_url ~* 'https?:\/\/.+'),
    CHECK(update_status_query IS NULL OR (update_status_query IS NOT NULL AND update_status_identifer IS NOT NULL))
);
