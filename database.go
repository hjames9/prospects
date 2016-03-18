package common

import (
	"database/sql"
	"fmt"
	"log"
)

type DatabaseCredentials struct {
	Driver       string
	Url          string
	User         string
	Password     string
	Name         string
	Host         string
	Port         string
	MaxOpenConns int
	MaxIdleConns int
}

func (dbCred DatabaseCredentials) IsValid() bool {
	result := false

	if len(dbCred.Driver) == 0 {
		result = false
	} else if len(dbCred.Url) > 0 {
		result = true
	} else if len(dbCred.User) > 0 && len(dbCred.Password) > 0 && len(dbCred.Name) > 0 {
		result = true
	}

	return result
}

func (dbCred DatabaseCredentials) GetString(useUrlArr ...bool) string {
	var dbInfo string

	useUrl := true
	if len(useUrlArr) > 0 {
		useUrl = useUrlArr[0]
	}

	if useUrl && len(dbCred.Url) > 0 {
		dbInfo = dbCred.Url
	} else {
		dbInfo = fmt.Sprintf("fallback_application_name=prospects user=%s password=%s dbname=%s host=%s port=%s", dbCred.User, dbCred.Password, dbCred.Name, dbCred.Host, dbCred.Port)
	}

	return dbInfo
}

func (dbCred DatabaseCredentials) GetDriver() string {
	return dbCred.Driver
}

func (dbCred DatabaseCredentials) GetDatabase() *sql.DB {
	db, err := sql.Open(dbCred.GetDriver(), dbCred.GetString())

	if nil != err {
		log.Printf("Error opening configured database: %s", dbCred.GetString())
		log.Print(err)
	} else {
		db.SetMaxOpenConns(dbCred.MaxOpenConns)
		db.SetMaxIdleConns(dbCred.MaxIdleConns)

		err = db.Ping()
		if nil != err {
			log.Printf("Error connecting to database: %s", dbCred.GetString())
			log.Print(err)
		}
	}

	return db
}
