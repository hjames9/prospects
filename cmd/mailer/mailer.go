package main

import (
	"database/sql"
	"flag"
	"github.com/hjames9/prospects"
	_ "github.com/lib/pq"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const (
	QUERY       = "SELECT mailer_name, source_email_address, get_email_data_query, dest_email_field_name, email_subject, email_subject_field_names, email_template_url, update_status_query, update_status_identifer FROM prospects.mailer_queries WHERE mailer_name = $1"
	LIMIT_REGEX = "LIMIT\\s+\\$1"
)

type MailerQuery struct {
	MailerName                string
	SourceEmailAddress        string
	GetEmailDataQuery         string
	DestinationEmailFieldName string
	EmailSubject              string
	EmailSubjectFieldNames    []string
	EmailTemplateUrl          string
	UpdateStatusQuery         string
	UpdateStatusIdentifer     string
}

func getMailerQuery(db *sql.DB, mailerName string) (MailerQuery, error) {
	var (
		mailerQuery            MailerQuery
		emailSubjectFieldNames sql.NullString
		updateStatusQuery      sql.NullString
		updateStatusIdentifer  sql.NullString
	)

	err := db.QueryRow(QUERY, mailerName).Scan(&mailerQuery.MailerName, &mailerQuery.SourceEmailAddress, &mailerQuery.GetEmailDataQuery, &mailerQuery.DestinationEmailFieldName, &mailerQuery.EmailSubject, &emailSubjectFieldNames, &mailerQuery.EmailTemplateUrl, &updateStatusQuery, &updateStatusIdentifer)

	if nil == err {
		if emailSubjectFieldNames.Valid {
			mailerQuery.EmailSubjectFieldNames = strings.Split(strings.Trim(emailSubjectFieldNames.String, "{}"), ",")
		}

		if updateStatusQuery.Valid {
			mailerQuery.UpdateStatusQuery = updateStatusQuery.String
		}

		if updateStatusIdentifer.Valid {
			mailerQuery.UpdateStatusIdentifer = updateStatusIdentifer.String
		}
	}

	return mailerQuery, err
}

type UpdateReplyStatus struct {
	query      string
	queryField string
	db         *sql.DB
	count      int
}

func (urs *UpdateReplyStatus) Processed(templateData map[string]string, completeTemplate string, success bool) bool {
	if success {
		_, err := urs.db.Exec(urs.query, time.Now(), templateData[urs.queryField])
		if nil != err {
			log.Print(err)
		}
		urs.count++
		return true
	} else {
		return false
	}
}

func main() {
	//SMTP server information
	smtpHost := os.Getenv("SMTP_HOST")
	smtpUser := os.Getenv("SMTP_USER")
	smtpPassword := os.Getenv("SMTP_PASSWORD")

	if len(smtpHost) <= 0 {
		log.Fatal("SMTP_HOST is NOT set")
	}

	if len(smtpUser) <= 0 {
		log.Fatal("SMTP_USER is NOT set")
	}

	if len(smtpPassword) <= 0 {
		log.Fatal("SMTP_PASSWORD is NOT set")
	}

	//Database connection
	dbUrl := os.Getenv("DATABASE_URL")
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")
	dbHost := common.GetenvWithDefault("DB_HOST", "localhost")
	dbPort := common.GetenvWithDefault("DB_PORT", "5432")
	dbMaxOpenConnsStr := common.GetenvWithDefault("DB_MAX_OPEN_CONNS", "10")
	dbMaxIdleConnsStr := common.GetenvWithDefault("DB_MAX_IDLE_CONNS", "0")

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

	//Get database connection
	log.Print("Enabling database connectivity")
	dbCredentials := common.DatabaseCredentials{common.DB_DRIVER, dbUrl, dbUser, dbPassword, dbName, dbHost, dbPort, dbMaxOpenConns, dbMaxIdleConns}
	if !dbCredentials.IsValid() {
		log.Fatalf("Database credentials NOT set correctly. %#v", dbCredentials)
	}
	db := dbCredentials.GetDatabase()
	defer db.Close()

	//Command line arguments
	mailerName := flag.String("mailer_name", "", "Name of mailer to process")
	processAmt := flag.Int("process_amt", 3, "Amount of mails to process")
	flag.Parse()

	if len(*mailerName) <= 0 {
		flag.Usage()
		os.Exit(1)
	}

	//Mailer query
	mailerQuery, err := getMailerQuery(db, *mailerName)
	if nil != err {
		log.Printf("Could not retrieve mailer query data for %s", *mailerName)
		log.Fatal(err)
	}

	//Limit regex
	log.Print("Compiling limit regular expression")
	limitRegex, err := regexp.Compile(LIMIT_REGEX)
	if nil != err {
		log.Print(err)
		log.Fatalf("Limit regex compilation failed for %s", LIMIT_REGEX)
	}

	dbParameters := make([]interface{}, 0)

	if limitRegex.MatchString(mailerQuery.GetEmailDataQuery) {
		log.Print("Email data query has parameterized LIMIT clause.  Adding process amount")
		dbParameters = append(dbParameters, *processAmt)
	} else {
		log.Print("Query doesn't contain LIMIT statement.  Process amount ignored")
	}

	urs := &UpdateReplyStatus{mailerQuery.UpdateStatusQuery, mailerQuery.UpdateStatusIdentifer, db, 0}

	emailSubjectFieldNames := make([]interface{}, 0)
	for _, emailSubjectFieldName := range mailerQuery.EmailSubjectFieldNames {
		emailSubjectFieldNames = append(emailSubjectFieldNames, emailSubjectFieldName)
	}

	dtm := common.DatabaseTemplateMailer{smtpHost, smtpUser, smtpPassword, mailerQuery.EmailSubject, emailSubjectFieldNames, mailerQuery.EmailTemplateUrl, mailerQuery.GetEmailDataQuery, dbParameters, mailerQuery.DestinationEmailFieldName, mailerQuery.SourceEmailAddress, db, urs}

	err = dtm.SendMail()
	if nil != err {
		log.Print("Error sending e-mails")
		log.Fatal(err)
	} else if urs.count == 0 {
		log.Print("No new messages to send")
	} else {
		log.Printf("Successfully sent %d e-mails", urs.count)
	}
}
