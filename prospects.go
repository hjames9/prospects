package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/go-martini/martini"
	_ "github.com/lib/pq"
	"github.com/martini-contrib/binding"
	"github.com/martini-contrib/cors"
	"log"
	"net/http"
	"os"
	"regexp"
	"time"
)

const (
	QUERY       = "INSERT INTO prospects (app_name, email, referrer, first_name, last_name, phone_number, created_at) VALUES($1, $2, $3, $4, $5, $6, $7) RETURNING id;"
	EMAIL_REGEX = "([\\w\\d\\.]+)@[\\w\\d\\.]+"
	POST_URL    = "/prospects"
	DB_DRIVER   = "postgres"
)

type ProspectForm struct {
	AppName     string `form:"appname" binding:"required"`
	Referrer    string `form:"referrer"`
	FirstName   string `form:"firstname"`
	LastName    string `form:"lastname"`
	Email       string `form:"email" binding:"required"`
	PhoneNumber string `form:"phonenumber"`
}

type Response struct {
	Code    int
	Message string
}

type CreateHandler func(http.ResponseWriter, ProspectForm) (int, string)
type NotFoundHandler func(http.ResponseWriter, *http.Request) (int, string)

func GetenvWithDefault(envKey string, defaultVal string) string {
	envVal := os.Getenv(envKey)

	if len(envVal) == 0 {
		envVal = defaultVal
	}

	return envVal
}

func main() {
	//Database connection
	log.Print("Enabling database connectivity")

	dbUrl := os.Getenv("DATABASE_URL")
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")
	dbHost := GetenvWithDefault("DB_HOST", "localhost")
	dbPort := GetenvWithDefault("DB_PORT", "5432")

	dbCredentials := DatabaseCredentials{DB_DRIVER, dbUrl, dbUser, dbPassword, dbName, dbHost, dbPort}
	if !dbCredentials.IsValid() {
		log.Fatalf("Database credentials NOT set correctly. %#v", dbCredentials)
	}

	db := dbCredentials.GetDatabase()
	defer db.Close()

	//Regular expression
	log.Print("Compiling e-mail regular expression")
	emailRegex, err := regexp.Compile(EMAIL_REGEX)
	if nil != err {
		log.Fatalf("Regex compilation failed for %s", EMAIL_REGEX)
	}

	//HTTP handlers
	log.Print("Preparing HTTP handlers")
	createHandler, notFoundHandler := setupHttpHandlers(db, emailRegex)

	//HTTP server
	host := GetenvWithDefault("HOST", "")
	port := GetenvWithDefault("PORT", "3000")
	mode := GetenvWithDefault("MARTINI_ENV", "development")

	log.Printf("Running HTTP server on %s:%s in mode %s", host, port, mode)
	runHttpServer(createHandler, notFoundHandler)
}

func setupHttpHandlers(db *sql.DB, emailRegex *regexp.Regexp) (CreateHandler, NotFoundHandler) {
	createHandler := func(res http.ResponseWriter, prospect ProspectForm) (int, string) {
		log.Printf("Received new prospect: %#v", prospect)

		res.Header().Set("Content-Type", "application/json")
		var response Response

		if emailRegex.MatchString(prospect.Email) {
			if nil != addProspect(db, prospect) {
				responseStr := fmt.Sprintf("Could not add prospect due to server error: e-mail %s", prospect.Email)
				response = Response{http.StatusInternalServerError, responseStr}
				log.Print(responseStr)
			} else {
				responseStr := fmt.Sprintf("Successfully added prospect. E-mail %s", prospect.Email)
				response = Response{http.StatusCreated, responseStr}
				log.Print(responseStr)
			}
		} else {
			responseStr := fmt.Sprintf("Could not add prospect. E-mail address %s is invalid", prospect.Email)
			response = Response{http.StatusBadRequest, responseStr}
			log.Print(responseStr)
		}

		jsonStr, _ := json.Marshal(response)
		return response.Code, string(jsonStr)
	}

	notFoundHandler := func(res http.ResponseWriter, req *http.Request) (int, string) {
		res.Header().Set("Content-Type", "application/json")
		responseStr := fmt.Sprintf("URL Not Found %s", req.URL)
		response := Response{http.StatusNotFound, responseStr}
		log.Print(responseStr)
		jsonStr, _ := json.Marshal(response)
		return response.Code, string(jsonStr)
	}

	return createHandler, notFoundHandler
}

func addProspect(db *sql.DB, prospect ProspectForm) error {
	var firstName sql.NullString
	if len(prospect.FirstName) != 0 {
		firstName = sql.NullString{prospect.FirstName, true}
	}

	var lastName sql.NullString
	if len(prospect.LastName) != 0 {
		lastName = sql.NullString{prospect.LastName, true}
	}

	var phoneNumber sql.NullString
	if len(prospect.PhoneNumber) != 0 {
		phoneNumber = sql.NullString{prospect.PhoneNumber, true}
	}

	var referrer sql.NullString
	if len(prospect.Referrer) != 0 {
		referrer = sql.NullString{prospect.Referrer, true}
	}

	var lastInsertId int
	err := db.QueryRow(QUERY, prospect.AppName, prospect.Email, referrer, firstName, lastName, phoneNumber, time.Now()).Scan(&lastInsertId)

	if nil == err {
		log.Printf("New prospect id = %d", lastInsertId)
	}

	return err
}

func runHttpServer(createHandler CreateHandler, notFoundHandler NotFoundHandler) {
	martini_ := martini.Classic()

	allowCORSHandler := cors.Allow(&cors.Options{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{"POST"},
		AllowHeaders: []string{"Origin"},
	})

	martini_.Post(POST_URL, allowCORSHandler, binding.Bind(ProspectForm{}), createHandler)
	martini_.NotFound(notFoundHandler)
	martini_.Run()
}
