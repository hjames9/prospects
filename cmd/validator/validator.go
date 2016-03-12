package main

import (
	"database/sql"
	"github.com/hjames9/prospects"
	_ "github.com/lib/pq"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
)

const (
	GET_LEADS_QUERY   = "SELECT id, lead_source, email, phone_number, miscellaneous, was_processed, is_valid FROM prospects.leads WHERE was_processed = FALSE ORDER BY id ASC LIMIT $1"
	UPDATE_LEAD_QUERY = "UPDATE prospects.leads SET was_processed = $1, is_valid = $2, miscellaneous = $3 WHERE id = $4"
	DB_DRIVER         = "postgres"
	USER_AGENT_HEADER = "User-Agent"
	USER_AGENT        = "Prospects validator"
)

func MakeHttpGetRequest(url string) ([]byte, int, error) {
	//Create HTTP client
	client := http.Client{}

	//Create request with headers
	request, err := http.NewRequest("GET", url, nil)
	request.Header.Add(USER_AGENT_HEADER, USER_AGENT)
	if nil != err {
		return nil, 0, err
	}

	//Execute request
	response, err := client.Do(request)
	if nil != err {
		return nil, 0, err
	}

	//Get response body
	body, err := ioutil.ReadAll(response.Body)
	response.Body.Close()
	if nil != err {
		return nil, 0, err
	}

	return body, response.StatusCode, nil
}

type Validator interface {
	Validate(Prospect) (bool, bool, string)
}

type Prospect struct {
	id            int
	leadSource    string
	email         sql.NullString
	phoneNumber   sql.NullString
	miscellaneous sql.NullString
	wasProcessed  bool
	isValid       bool
	validators    []Validator
}

func (prospect *Prospect) IsProcessed() bool {
	var masterMisc string

	for _, validator := range prospect.validators {
		isValid, wasProcessed, miscellaneous := validator.Validate(*prospect)
		if isValid {
			prospect.isValid = isValid
		}

		if wasProcessed {
			prospect.wasProcessed = wasProcessed
		}

		if masterMisc != "" && miscellaneous != "" {
			masterMisc += ", "
		}

		if miscellaneous != "" {
			masterMisc += miscellaneous
		}
	}

	if masterMisc != "" {
		prospect.miscellaneous.String = "[" + masterMisc + "]"
		prospect.miscellaneous.Valid = true
	}

	return prospect.wasProcessed
}

func GetenvWithDefault(envKey string, defaultVal string) string {
	envVal := os.Getenv(envKey)

	if len(envVal) == 0 {
		envVal = defaultVal
	}

	return envVal
}

func process(db *sql.DB, prospects []Prospect) {
	transaction, err := db.Begin()
	if nil != err {
		log.Print("Error creating transaction")
		log.Print(err)
	}

	defer transaction.Rollback()
	statement, err := transaction.Prepare(UPDATE_LEAD_QUERY)
	if nil != err {
		log.Print("Error preparing SQL statement")
		log.Print(err)
	}

	defer statement.Close()

	counter := 0
	unused := -1
	for _, prospect := range prospects {
		if prospect.IsProcessed() {
			err = statement.QueryRow(prospect.wasProcessed, prospect.isValid, prospect.miscellaneous, prospect.id).Scan(&unused)
			counter++
		}
	}

	err = transaction.Commit()
	if nil != err {
		log.Print("Error committing transaction")
		log.Print(err)
	} else {
		log.Printf("Processed %d prospects", counter)
	}
}

func main() {
	dbUrl := os.Getenv("DATABASE_URL")
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")
	dbHost := GetenvWithDefault("DB_HOST", "localhost")
	dbPort := GetenvWithDefault("DB_PORT", "5432")
	dbMaxOpenConnsStr := GetenvWithDefault("DB_MAX_OPEN_CONNS", "10")
	dbMaxIdleConnsStr := GetenvWithDefault("DB_MAX_IDLE_CONNS", "0")
	processAmtStr := GetenvWithDefault("PROCESS_AMT", "3")
	fullContactApiKey := os.Getenv("FULLCONTACT_APIKEY")
	numVerifyApiKey := os.Getenv("NUMVERIFY_APIKEY")

	if len(fullContactApiKey) <= 0 {
		log.Fatal("FullContact API key not set")
	}

	if len(numVerifyApiKey) <= 0 {
		log.Fatal("NumVerify API key not set")
	}

	dbMaxOpenConns, err := strconv.Atoi(dbMaxOpenConnsStr)
	if nil != err {
		dbMaxOpenConns = 10
		log.Printf("Error setting database maximum open connections from value: %s. Default to %d", dbMaxOpenConnsStr, dbMaxOpenConns)
		log.Print(err)
	}

	dbMaxIdleConns, err := strconv.Atoi(dbMaxIdleConnsStr)
	if nil != err {
		dbMaxIdleConns = 0
		log.Printf("Error setting database maximum idle connections from value: %s. Default to %d", dbMaxIdleConnsStr, dbMaxIdleConns)
		log.Print(err)
	}

	processAmt, err := strconv.Atoi(processAmtStr)
	if nil != err {
		processAmt = 3
		log.Printf("Error setting process amount from value: %s. Default to %d", processAmtStr, processAmt)
		log.Print(err)
	}

	//Database connection
	log.Print("Enabling database connectivity")

	dbCredentials := database.DatabaseCredentials{DB_DRIVER, dbUrl, dbUser, dbPassword, dbName, dbHost, dbPort, dbMaxOpenConns, dbMaxIdleConns}
	if !dbCredentials.IsValid() {
		log.Fatalf("Database credentials NOT set correctly. %#v", dbCredentials)
	}

	db := dbCredentials.GetDatabase()
	defer db.Close()

	rows, err := db.Query(GET_LEADS_QUERY, processAmt)
	if nil != err {
		log.Fatal(err)
	}
	defer rows.Close()

	var (
		id            int
		leadSource    string
		email         sql.NullString
		phoneNumber   sql.NullString
		miscellaneous sql.NullString
		wasProcessed  bool
		isValid       bool
	)

	var prospects []Prospect
	var validators []Validator
	validators = append(validators, FullContactValidator{fullContactApiKey})
	validators = append(validators, NumVerifyValidator{numVerifyApiKey})

	for rows.Next() {
		err := rows.Scan(&id, &leadSource, &email, &phoneNumber, &miscellaneous, &wasProcessed, &isValid)
		if nil != err {
			log.Fatal(err)
		}

		prospect := Prospect{id, leadSource, email, phoneNumber, miscellaneous, wasProcessed, isValid, validators}
		prospects = append(prospects, prospect)
	}

	err = rows.Err()
	if nil != err {
		log.Fatal(err)
	}

	process(db, prospects)
}
