SET search_path TO prospects,public;

COMMENT ON SCHEMA prospects IS 'Prospects schema holds all objects for application';

COMMENT ON TYPE gender IS 'Gender type between male or female';

COMMENT ON TYPE lead_source IS 'Source lead was generated from';

COMMENT ON TABLE leads IS 'Leads table provides unnormalized data for every data point a potential customer is willing to provide. A lead can use multiple rows to provide different data, depending on the interface workflow they''ve chosen.';

COMMENT ON COLUMN leads.id IS 'Primary key id of the current lead''s interaction.';
COMMENT ON COLUMN leads.lead_id IS 'Unique id (uuid) of the lead.';
COMMENT ON COLUMN leads.app_name IS 'Application name that lead is for.';
COMMENT ON COLUMN leads.email IS 'E-mail address of the lead.';
COMMENT ON COLUMN leads.lead_source IS 'Source lead was generated from.';
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
COMMENT ON COLUMN leads.replied_to IS 'Determines if lead was replied to or not.';
COMMENT ON COLUMN leads.created_at IS 'Timestamp of lead creation.';
COMMENT ON COLUMN leads.updated_at IS 'Timestamp of last time lead was updated.';

COMMENT ON CONSTRAINT leads_pkey ON leads IS 'Primary key constraint for leads id column.';
COMMENT ON CONSTRAINT leads_check ON leads IS 'Check constraint used to enforce that a given lead with a landing source has an e-mail address or phone number.';
COMMENT ON CONSTRAINT leads_check1 ON leads IS 'Check constraint used to enforce that a given lead with an email source has an e-mail address.';
COMMENT ON CONSTRAINT leads_check2 ON leads IS 'Check constraint used to enforce that a given lead with a phone source has a phone number.';
COMMENT ON CONSTRAINT leads_check3 ON leads IS 'Check constraint used to enforce that a given lead with a feedback source has feedback.';
COMMENT ON CONSTRAINT leads_check4 ON leads IS 'Check constraint used to enforce that a given lead with a extended source has an extended field.';
COMMENT ON CONSTRAINT leads_email_check ON leads IS 'Check constraint used to enforce correct e-mail address format.';
COMMENT ON CONSTRAINT leads_geolocation_check ON leads IS 'Check constraint used to enforce correct values for latitude and longitude.';

COMMENT ON SEQUENCE leads_id_seq IS 'Primary key sequence for leads table.  Values are obfuscated since they''re used on public interfaces';

COMMENT ON VIEW sneezers IS 'Sneezers attempt to aggregate relevant data, for a particular lead, from the leads table, into a single row';

COMMENT ON RULE "_RETURN" ON sneezers IS 'Internal rule for sneezers view';

COMMENT ON COLUMN sneezers.id IS 'Primary key of current lead''s interaction';
COMMENT ON COLUMN sneezers.lead_id IS 'Unique id (uuid) of sneezer';
COMMENT ON COLUMN sneezers.app_name IS 'Application name that sneezer is accessing';
COMMENT ON COLUMN sneezers.email IS 'E-mail address of sneezer';
COMMENT ON COLUMN sneezers.lead_source IS 'Source lead was generated from';
COMMENT ON COLUMN sneezers.feedback IS 'Feedback provided by sneezer';
COMMENT ON COLUMN sneezers.first_name IS 'First name of sneezer';
COMMENT ON COLUMN sneezers.last_name IS 'Last name of sneezer';
COMMENT ON COLUMN sneezers.phone_number IS 'Phone number of sneezer';
COMMENT ON COLUMN sneezers.dob IS 'Date of birth of sneezer';
COMMENT ON COLUMN sneezers.gender IS 'Gender of sneezer';
COMMENT ON COLUMN sneezers.zip_code IS 'Zip code of sneezer';
COMMENT ON COLUMN sneezers.language IS 'Language setting of sneezer';
COMMENT ON COLUMN sneezers.user_agent IS 'User agent of web browser used by sneezer';
COMMENT ON COLUMN sneezers.miscellaneous IS 'Adhoc miscellaneous data that can be provided.';
COMMENT ON COLUMN sneezers.was_processed IS 'Determines if sneezer information verification was attempted or not.';
COMMENT ON COLUMN sneezers.is_valid IS 'Determines if sneezer was determined to be valid or not.';
COMMENT ON COLUMN sneezers.replied_to IS 'Determines if sneezer was replied to or not.';
COMMENT ON COLUMN sneezers.created_at IS 'Timestamp of sneezer creation';
COMMENT ON COLUMN sneezers.updated_at IS 'Timestamp of last time sneezer was updated.';

COMMENT ON INDEX l_lead_id_idx IS 'Index for unique id generated by lead, that can span multiple rows.  This helps for querying all the data points that a lead has provided';
COMMENT ON INDEX l_app_name_idx IS 'Index for application name.  This helps for querying all data points for a particular application';
COMMENT ON INDEX l_email_idx IS 'Index for lead e-mail addresses.  This helps querying all data points for a particular e-mail address';
COMMENT ON INDEX l_referrer_idx IS 'Index for page referrers. This helps querying all data points for the web pages that referred us to a particular landing page containing the interacting form.';
COMMENT ON INDEX l_misc_idx IS 'Index for miscellaneous jsonb field.  This will allow for any future potential data we want to add that isn''t currently modeled but yet we would want to search for.';

COMMENT ON TABLE imap_markers IS 'Table is used to track prospects received via e-mail';
COMMENT ON COLUMN imap_markers.app_name IS 'Application name that lead is for.';
COMMENT ON COLUMN imap_markers.marker IS 'Current position in imap mailbox';
COMMENT ON COLUMN imap_markers.updated_at IS 'Time of last mailbox position change';
COMMENT ON CONSTRAINT imap_markers_pkey ON imap_markers IS 'Primary key constraint for imap_markers app_name column.';
COMMENT ON CONSTRAINT imap_markers_marker_check ON imap_markers IS 'Check constraint used to enforce that a marker is at least 1.';
