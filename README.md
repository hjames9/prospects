# Prospects:  Toolkit for collecting sales prospects
A Go-based http server and e-mail processor to collect potential prospects and persist them to a Postgresql database.  Also provides validation of collected data and HTML e-mail responses.

## prospects - http server

### Setup - Set environmental variables
    DATABASE_URL=postgres://user:password@localhost:5432/prospect_db (no default)
    DB_USER=hjames (no default, ignored with DATABASE_URL set)
    DB_PASSWORD=blahblah (no default, ignored with DATABASE_URL set)
    DB_NAME=prospect_db (no default, ignored with DATABASE_URL set)
    DB_HOST=localhost (default is localhost, ignored with DATABASE_URL set)
    DB_PORT=5432 (default is 5432, ignored with DATABASE_URL set)
    DB_MAX_OPEN_CONNS=100 (default is 10)
    DB_MAX_IDLE_CONNS=100 (default is 0)
    PGAPPNAME=prospects (default is prospects)
    SSL_REDIRECT=true (default is false)
    GZIP_RESPONSE=false (default is true)
    GZIP_COMPRESSION_LEVEL=9 (Any value 1-9, default is 6)
    HOST=localhost (default is all interfaces (blank))
    PORT=8080 (default is 3000)
    MARTINI_ENV=production (default is development)
    APPLICATION_NAMES=tremont,laconia,paulding (default is empty for all allowable application names)
    ALLOW_HEADERS=X-Requested-With,X-Forwarded-For (default is empty for only default headers)
    BOTDETECT_FIELDLOCATION=body (default is body, can be body or header)
    BOTDETECT_FIELDNAME=middlename (default is spambot)
    BOTDETECT_FIELDVALUE=iamhuman (default is blank)
    BOTDETECT_MUSTMATCH=true (default is true)
    BOTDETECT_PLAYCOY=true (default is true)
    ASYNC_REQUEST=true (default is false)
    ASYNC_REQUEST_SIZE=100000 (default is 100000)
    ASYNC_PROCESS_INTERVAL=10 (default is 5 seconds)
    IP_ADDRESS_LOCATION=xff_first (default is normal, can be normal, xff_first, xff_last)
    STRING_SIZE_LIMIT=1000 (default is 500)
    FEEDBACK_SIZE_LIMIT=5000 (default is 3000)

## emissary - e-mail prospects retriever

### Setup - Set environmental variables
    DATABASE_URL=postgres://user:password@localhost:5432/prospect_db (no default)
    DB_USER=hjames (no default, ignored with DATABASE_URL set)
    DB_PASSWORD=blahblah (no default, ignored with DATABASE_URL set)
    DB_NAME=prospect_db (no default, ignored with DATABASE_URL set)
    DB_HOST=localhost (default is localhost, ignored with DATABASE_URL set)
    DB_PORT=5432 (default is 5432, ignored with DATABASE_URL set)
    DB_MAX_OPEN_CONNS=100 (default is 10)
    DB_MAX_IDLE_CONNS=100 (default is 0)
    APPLICATION_NAME=tremont (no default)
    IMAPS_HOST=imap.gmail.com:993 (no default)
    IMAPS_USER=info@best_products.com (no default)
    IMAPS_PASSWORD=blahblah (no default)
    IMAPS_MAILBOX=SPECIAL (default is "INBOX")

## validator - data validation

### Setup - Set environmental variables
    DATABASE_URL=postgres://user:password@localhost:5432/prospect_db (no default)
    DB_USER=hjames (no default, ignored with DATABASE_URL set)
    DB_PASSWORD=blahblah (no default, ignored with DATABASE_URL set)
    DB_NAME=prospect_db (no default, ignored with DATABASE_URL set)
    DB_HOST=localhost (default is localhost, ignored with DATABASE_URL set)
    DB_PORT=5432 (default is 5432, ignored with DATABASE_URL set)
    DB_MAX_OPEN_CONNS=100 (default is 10)
    DB_MAX_IDLE_CONNS=100 (default is 0)
    PROCESS_AMT=3 (default is 3)
    FULLCONTACT_APIKEY=0d9817d9-b9bd-4e15-871b-a2a3a1101ab5 (no default)
    NUMVERIFY_APIKEY=d7f10b5a-e34d-4c75-8345-425691939c36 (no default)

## mailer - e-mail responses to prospects

### Setup - Set environmental variables
    DATABASE_URL=postgres://user:password@localhost:5432/prospect_db (no default)
    DB_USER=hjames (no default, ignored with DATABASE_URL set)
    DB_PASSWORD=blahblah (no default, ignored with DATABASE_URL set)
    DB_NAME=prospect_db (no default, ignored with DATABASE_URL set)
    DB_HOST=localhost (default is localhost, ignored with DATABASE_URL set)
    DB_PORT=5432 (default is 5432, ignored with DATABASE_URL set)
    DB_MAX_OPEN_CONNS=100 (default is 10)
    DB_MAX_IDLE_CONNS=100 (default is 0)
    PROCESS_AMT=3 (default is 3)
    SMTP_HOST=smtp.gmail.com:587 (no default)
    SMTP_USER=info@best_products.com (no default)
    SMTP_PASSWORD=blahblah (no default)
    SMTP_REPLY_TEMPLATE_URL=http://dev.best_products.com/email.html (no default)
    SMTP_REPLY_SUBJECT=Thank you your interest! (no default)
