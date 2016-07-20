package main

import (
	"bitbucket.org/savewithus/prospects"
	"database/sql"
	_ "github.com/lib/pq"
	"log"
	"os"
	"strconv"
	"time"
)

const (
	FROM_QUERY        = "FROM prospects.leads WHERE was_processed = FALSE ORDER BY id ASC LIMIT $1"
	UPDATE_LEAD_QUERY = "UPDATE prospects.leads SET was_processed = $1, is_valid = $2, miscellaneous = miscellaneous || $3, updated_at = $4 WHERE id = $5"
)

type Validator interface {
	Validate(common.Prospect) (bool, bool, string)
}

func IsProcessed(prospect *common.Prospect, validators []Validator) bool {
	var masterMisc string

	for _, validator := range validators {
		isValid, wasProcessed, miscellaneous := validator.Validate(*prospect)
		if isValid {
			prospect.IsValid = isValid
		}

		if wasProcessed {
			prospect.WasProcessed = wasProcessed
		}

		if masterMisc != "" && miscellaneous != "" {
			masterMisc += ", "
		}

		if miscellaneous != "" {
			masterMisc += miscellaneous
		}
	}

	if masterMisc != "" {
		prospect.Miscellaneous = "[" + masterMisc + "]"
	}

	return prospect.WasProcessed
}

func process(db *sql.DB, prospects []common.Prospect, validators []Validator) {
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
		if IsProcessed(&prospect, validators) {
			err = statement.QueryRow(prospect.WasProcessed, prospect.IsValid, prospect.Miscellaneous, time.Now(), prospect.Id).Scan(&unused)
			if nil != err && sql.ErrNoRows != err {
				log.Printf("Error processing %#v", prospect)
				log.Print(err)
			}
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
	dbHost := common.GetenvWithDefault("DB_HOST", "localhost")
	dbPort := common.GetenvWithDefault("DB_PORT", "5432")
	dbMaxOpenConnsStr := common.GetenvWithDefault("DB_MAX_OPEN_CONNS", "10")
	dbMaxIdleConnsStr := common.GetenvWithDefault("DB_MAX_IDLE_CONNS", "0")
	processAmtStr := common.GetenvWithDefault("PROCESS_AMT", "3")
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

	dbCredentials := common.DatabaseCredentials{common.DB_DRIVER, dbUrl, dbUser, dbPassword, dbName, dbHost, dbPort, dbMaxOpenConns, dbMaxIdleConns}
	if !dbCredentials.IsValid() {
		log.Fatalf("Database credentials NOT set correctly. %#v", dbCredentials)
	}

	db := dbCredentials.GetDatabase()
	defer db.Close()

	prospects, err := common.GetProspects(db, FROM_QUERY, processAmt)
	if nil != err {
		log.Fatal(err)
	} else {
		log.Printf("Successfully fetched %d prospects", len(prospects))
	}

	var validators []Validator
	validators = append(validators, FullContactValidator{fullContactApiKey})
	validators = append(validators, NumVerifyValidator{numVerifyApiKey})

	process(db, prospects, validators)
}
