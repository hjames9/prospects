# prospects
A Go-based http server to collect potential prospects and persist them to a Postgresql database

## Setup - Set environmental variables
    DATABASE_URL=postgres://user:password@localhost:5432/prospect_db (no default)
    DB_USER=hjames (no default, ignored with DATABASE_URL set)
    DB_PASSWORD=blahblah (no default, ignored with DATABASE_URL set)
    DB_NAME=prospect_db (no default, ignored with DATABASE_URL set)
    DB_HOST=localhost (default is localhost)
    DB_PORT=5432 (default is 5432)
    HOST=localhost (default is all interfaces (blank))
    PORT=8080 (default is 3000)
    MARTINI_ENV=production (default is development)
