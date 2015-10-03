# prospects
A Go-based http server to collect potential prospects and persist them to a Postgresql database

## Setup - Set environmental variables
    DATABASE_URL=postgres://user:password@localhost:5432/prospect_db (no default)
    DB_USER=hjames (no default, ignored with DATABASE_URL set)
    DB_PASSWORD=blahblah (no default, ignored with DATABASE_URL set)
    DB_NAME=prospect_db (no default, ignored with DATABASE_URL set)
    DB_HOST=localhost (default is localhost, ignored with DATABASE_URL set)
    DB_PORT=5432 (default is 5432, ignored with DATABASE_URL set)
    DB_MAX_OPEN_CONNS=100 (default is 10)
    SSL_REDIRECT=true (default is false)
    HOST=localhost (default is all interfaces (blank))
    PORT=8080 (default is 3000)
    MARTINI_ENV=production (default is development)
    APPLICATION_NAMES=tremont,laconia,paulding (default is empty for all allowable application names)
