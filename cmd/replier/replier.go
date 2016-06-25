package main

import (
	"bytes"
	"database/sql"
	"github.com/hjames9/prospects"
	_ "github.com/lib/pq"
	"gopkg.in/gomail.v2"
	"html/template"
	"log"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	FROM_QUERY        = "FROM prospects.sneezers WHERE lead_source IN ('email', 'landing') AND email IS NOT NULL AND replied_to = FALSE ORDER BY id ASC LIMIT $1;"
	UPDATE_QUERY      = "UPDATE prospects.leads SET replied_to = TRUE, updated_at = $1 WHERE id = $2"
	TO_HEADER         = "To"
	SUBJECT_HEADER    = "Subject"
	HTML_CONTENT_TYPE = "text/html"
)

func sendEmailReply(smtpServer string, smtpUser string, smtpPassword string, smtpReplyTemplateUrl *url.URL, smtpReplySubject string, db *sql.DB, prospects []common.Prospect) error {
	//Get HTML template
	smtpReplyTemplate, responseCode, _, err := common.MakeHttpGetRequest(smtpReplyTemplateUrl.String())
	if nil != err {
		return err
	}

	if responseCode >= 200 && responseCode <= 299 && len(smtpReplyTemplate) > 0 {
		//Connect to smtp server
		smtpPair := strings.Split(smtpServer, ":")
		smtpPort := 25
		if len(smtpPair) == 2 {
			smtpPort, err = strconv.Atoi(smtpPair[1])
			smtpServer = smtpPair[0]
			if nil != err {
				log.Printf("Invalid port number specified: %s.  Setting to default port 25.", smtpPair[1])
				log.Print(err)
				smtpPort = 25
			}
		}

		//HTML templating
		tmpl, err := template.New("foo").Parse(string(smtpReplyTemplate))
		if nil != err {
			return err
		}
		var tmplBuffer bytes.Buffer

		//SMTP client
		smtpClient := gomail.NewDialer(smtpServer, smtpPort, smtpUser, smtpPassword)
		sender, err := smtpClient.Dial()
		if nil != err {
			return err
		}
		defer sender.Close()

		for _, prospect := range prospects {
			err = tmpl.Execute(&tmplBuffer, prospect)
			if nil != err {
				return err
			}

			message := gomail.NewMessage()
			message.SetHeader(common.FROM_HEADER, smtpUser)
			message.SetHeader(TO_HEADER, prospect.Email)
			message.SetHeader(SUBJECT_HEADER, smtpReplySubject)
			message.SetHeader(common.USER_AGENT_HEADER, common.USER_AGENT)
			message.SetBody(HTML_CONTENT_TYPE, tmplBuffer.String())

			err = sender.Send(smtpUser, []string{prospect.Email}, message)
			if nil != err {
				log.Print(err)
			} else {
				_, err := db.Exec(UPDATE_QUERY, time.Now(), prospect.Id)
				if nil != err {
					log.Print(err)
				}
			}
		}
	}

	return nil
}

func main() {
	//SMTP server and reply template
	smtpHost := os.Getenv("SMTP_HOST")
	smtpUser := os.Getenv("SMTP_USER")
	smtpPassword := os.Getenv("SMTP_PASSWORD")
	smtpReplyTemplateUrlStr := os.Getenv("SMTP_REPLY_TEMPLATE_URL")
	smtpReplySubject := os.Getenv("SMTP_REPLY_SUBJECT")

	if len(smtpHost) <= 0 {
		log.Fatal("SMTP_HOST is NOT set")
	}

	if len(smtpUser) <= 0 {
		log.Fatal("SMTP_USER is NOT set")
	}

	if len(smtpPassword) <= 0 {
		log.Fatal("SMTP_PASSWORD is NOT set")
	}

	if len(smtpReplyTemplateUrlStr) <= 0 {
		log.Fatal("SMTP_REPLY_TEMPLATE_URL is NOT set")
	}

	smtpReplyTemplateUrl, err := url.Parse(smtpReplyTemplateUrlStr)
	if nil != err {
		log.Printf("SMTP reply template URL is invalid: %s", smtpReplyTemplateUrlStr)
		log.Fatal(err)
	}

	if len(smtpReplySubject) <= 0 {
		log.Fatal("SMTP_REPLY_SUBJECT is NOT set")
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
	processAmtStr := common.GetenvWithDefault("PROCESS_AMT", "3")

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

	//Get process amount
	processAmt, err := strconv.Atoi(processAmtStr)
	if nil != err {
		processAmt = 3
		log.Printf("Error setting process amount from value: %s. Default to %d", processAmtStr, processAmt)
		log.Print(err)
	}

	//Get latest prospects
	log.Print("Fetching latest prospects")
	prospects, err := common.GetProspects(db, FROM_QUERY, processAmt)
	if nil != err {
		log.Fatal(err)
	} else {
		log.Printf("Successfully fetched %d prospects", len(prospects))
	}

	//Send thank you reply
	err = sendEmailReply(smtpHost, smtpUser, smtpPassword, smtpReplyTemplateUrl, smtpReplySubject, db, prospects)
	if nil != err {
		log.Print("Error sending e-mails")
		log.Fatal(err)
	} else if len(prospects) == 0 {
		log.Print("No new prospects received")
	} else {
		log.Print("Successfully sent e-mails")
	}
}
