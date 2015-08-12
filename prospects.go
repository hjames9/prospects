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
	QUERY       = "INSERT INTO prospects (app_name, email, first_name, last_name, created_at) VALUES($1, $2, $3, $4, $5) RETURNING id;"
	EMAIL_REGEX = "([\\w\\d\\.]+)@[\\w\\d\\.]+"
	POST_URL    = "/prospects"
)

type ProspectForm struct {
	AppName   string `form:"appname" binding:"required"`
	FirstName string `form:"firstname"`
	LastName  string `form:"lastname"`
	Email     string `form:"email" binding:"required"`
}

type Response struct {
	Code    int
	Message string
}

type CreateHandler func(http.ResponseWriter, ProspectForm) (int, string)
type NotFoundHandler func(http.ResponseWriter, *http.Request) (int, string)

func main() {
	//Database connection
	log.Print("Enabling database connectivity")

	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")

	if len(user) == 0 || len(password) == 0 || len(dbName) == 0 {
		log.Fatalf("Database credentials NOT set correctly. User(%s), Password(%s), Database name(%s)", user, password, dbName)
	}

	db := setupDatabase(user, password, dbName)
	defer db.Close()

	//Regular expression
	log.Print("Compiling e-mail regular expression")
	emailRegex, _ := regexp.Compile(EMAIL_REGEX)

	//HTTP handlers
	log.Print("Preparing HTTP handlers")
	createHandler, notFoundHandler := setupHttpHandlers(db, emailRegex)

	//HTTP server
	host := os.Getenv("HTTP_HOST")
	if len(host) == 0 {
		host = ""
	}

	port := os.Getenv("HTTP_PORT")
	if len(port) == 0 {
		port = "3000"
	}

	mode := os.Getenv("MARTINI_ENV")
	if len(mode) == 0 {
		mode = "development"
	}

	log.Printf("Running HTTP server on %s:%s in mode %s", host, port, mode)
	runHttpServer(createHandler, notFoundHandler, host, port)
}

func setupDatabase(user string, password string, dbName string) *sql.DB {
	dbinfo := fmt.Sprintf("user=%s password=%s dbname=%s sslmode=disable", user, password, dbName)

	db, err := sql.Open("postgres", dbinfo)
	if nil != err {
		log.Print("Error opening configured database")
	}

	err = db.Ping()
	if nil != err {
		log.Print("Error connecting to database")
	}

	return db
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

	var lastInsertId int
	err := db.QueryRow(QUERY, prospect.AppName, prospect.Email, firstName, lastName, time.Now()).Scan(&lastInsertId)

	if nil == err {
		log.Printf("New prospect id = %d", lastInsertId)
	}

	return err
}

func runHttpServer(createHandler CreateHandler, notFoundHandler NotFoundHandler, host string, port string) {
	os.Setenv("HOST", host)
	os.Setenv("PORT", port)

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
