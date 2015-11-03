SET search_path TO prospects,public;

COMMENT ON SCHEMA prospects IS 'Prospects schema holds all objects for application';

COMMENT ON TYPE gender IS 'Gender type between male or female';

COMMENT ON TABLE leads IS 'Leads table provides unnormalized data for every data point a potential customer is willing to provide. A lead can use multiple rows to provide different data, depending on the interface workflow they''ve chosen.';

COMMENT ON COLUMN leads.id IS 'Primary key id of the current lead''s interaction.';
COMMENT ON COLUMN leads.lead_id IS 'Unique id (uuid) of the lead.';
COMMENT ON COLUMN leads.app_name IS 'Application name that lead is for.';
COMMENT ON COLUMN leads.email IS 'E-mail address of the lead.';
COMMENT ON COLUMN leads.used_pinterest IS 'Whether lead came from pinterest.';
COMMENT ON COLUMN leads.used_facebook IS 'Whether lead came from facebook.';
COMMENT ON COLUMN leads.used_instagram IS 'Whether lead came from instagram.';
COMMENT ON COLUMN leads.used_twitter IS 'Whether lead came from twitter.';
COMMENT ON COLUMN leads.used_google IS 'Whether lead came from google.';
COMMENT ON COLUMN leads.used_youtube IS 'Whether lead came from youtube.';
COMMENT ON COLUMN leads.extended IS 'Extended data provided by lead.';
COMMENT ON COLUMN leads.feedback IS 'Feedback provided by lead.';
COMMENT ON COLUMN leads.first_name IS 'First name of lead.';
COMMENT ON COLUMN leads.last_name IS 'Last name of lead.';
COMMENT ON COLUMN leads.phone_number IS 'Phone number of lead.';
COMMENT ON COLUMN leads.gender IS 'Gender of lead.';
COMMENT ON COLUMN leads.language IS 'Language setting of lead.';
COMMENT ON COLUMN leads.dob IS 'Date of birth of the lead.';
COMMENT ON COLUMN leads.zip_code IS 'Zip code of the lead.';
COMMENT ON COLUMN leads.referrer IS 'Landing page that contained form.';
COMMENT ON COLUMN leads.page_referrer IS 'Page that referred landing page.';
COMMENT ON COLUMN leads.user_agent IS 'User agent of web browser accessing form.';
COMMENT ON COLUMN leads.cookies IS 'Available cookies of lead.';
COMMENT ON COLUMN leads.geolocation IS 'Latitude and Longtiude of lead.';
COMMENT ON COLUMN leads.ip_address IS 'IP address of lead.';
COMMENT ON COLUMN leads.miscellaneous IS 'Adhoc miscellaneous data that can be provided.';
COMMENT ON COLUMN leads.was_processed IS 'Determines if lead information verification was attempted or not.';
COMMENT ON COLUMN leads.is_valid IS 'Determines if lead was determined to be valid or not.';
COMMENT ON COLUMN leads.created_at IS 'Timestamp of lead creation.';

COMMENT ON CONSTRAINT leads_pkey ON leads IS 'Primary key constraint for leads id column.';
COMMENT ON CONSTRAINT leads_check ON leads IS 'Check constraint used to enforce that a given lead provides source information via either e-mail address, social media reference or feedback.';
COMMENT ON CONSTRAINT leads_email_check ON leads IS 'Check constraint used to enforce correct e-mail address format.';
COMMENT ON CONSTRAINT leads_geolocation_check ON leads IS 'Check constraint used to enforce correct values for latitude and longtiude.';

COMMENT ON SEQUENCE leads_id_seq IS 'Primary key sequence for leads table.  Values are obfuscated since they''re used on public interfaces';

COMMENT ON VIEW sneezers IS 'Sneezers attempt to aggregate relevant data, for a particular lead, from the leads table, into a single row';

COMMENT ON RULE "_RETURN" ON sneezers IS 'Internal rule for sneezers view';

COMMENT ON COLUMN sneezers.id IS 'Primary key of current lead''s interaction';
COMMENT ON COLUMN sneezers.lead_id IS 'Unique id (uuid) of sneezer';
COMMENT ON COLUMN sneezers.app_name IS 'Application name that sneezer is accessing';
COMMENT ON COLUMN sneezers.email IS 'E-mail address of sneezer';
COMMENT ON COLUMN sneezers.used_pinterest IS 'Whether sneezer came from pinterest';
COMMENT ON COLUMN sneezers.used_facebook IS 'Whether sneezer came from facebook';
COMMENT ON COLUMN sneezers.used_instagram IS 'Whether sneezer came from instagram';
COMMENT ON COLUMN sneezers.used_twitter IS 'Whether sneezer came from twitter';
COMMENT ON COLUMN sneezers.used_google IS 'Whether sneezer came from google';
COMMENT ON COLUMN sneezers.used_youtube IS 'Whether sneezer came from youtube';
COMMENT ON COLUMN sneezers.extended IS 'Extended data provided by sneezer.';
COMMENT ON COLUMN sneezers.feedback IS 'Feedback provided by sneezer';
COMMENT ON COLUMN sneezers.first_name IS 'First name of sneezer';
COMMENT ON COLUMN sneezers.last_name IS 'Last name of sneezer';
COMMENT ON COLUMN sneezers.phone_number IS 'Phone number of sneezer';
COMMENT ON COLUMN sneezers.dob IS 'Date of birth of sneezer';
COMMENT ON COLUMN sneezers.gender IS 'Gender of sneezer';
COMMENT ON COLUMN sneezers.zip_code IS 'Zip code of sneezer';
COMMENT ON COLUMN sneezers.language IS 'Language setting of sneezer';
COMMENT ON COLUMN sneezers.user_agent IS 'User agent of web browser used by sneezer';
COMMENT ON COLUMN sneezers.created_at IS 'Timestamp of sneezer creation';

COMMENT ON INDEX l_lead_id_idx IS 'Index for unique id generated by lead, that can span multiple rows.  This helps for querying all the data points that a lead has provided';
COMMENT ON INDEX l_app_name_idx IS 'Index for application name.  This helps for querying all data points for a particular application';
COMMENT ON INDEX l_email_idx IS 'Index for lead e-mail addresses.  This helps querying all data points for a particular e-mail address';
COMMENT ON INDEX l_referrer_idx IS 'Index for page referrers. This helps querying all data points for the web pages that referred us to a particular landing page containing the interacting form.';
COMMENT ON INDEX l_misc_idx IS 'Index for miscellaneous jsonb field.  This will allow for any future potential data we want to add that isn''t currently modeled but yet we would want to search for.';
