package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"github.com/hjames9/prospects"
	_ "github.com/lib/pq"
	"github.com/mxk/go-imap/imap"
	"github.com/satori/go.uuid"
	"gopkg.in/gomail.v2"
	"log"
	"net/mail"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const (
	QUERY           = "INSERT INTO prospects.leads(lead_id, app_name, lead_source, email, user_agent, miscellaneous, created_at) VALUES($1, $2, $3, $4, $5, $6, $7) RETURNING id;"
	GET_IMAP_MARKER = "SELECT last_value FROM prospects.imap_marker"
	SET_IMAP_MARKER = "SELECT setval('prospects.imap_marker', $1)"
	DB_DRIVER       = "postgres"
	EMAIL_REGEX     = "[A-Za-z0-9._%-]+@[A-Za-z0-9.-]+[.][A-Za-z]+"
)

type Prospect struct {
	LeadId        uuid.UUID
	AppName       string
	LeadSource    string
	Email         string
	UserAgent     sql.NullString
	Miscellaneous sql.NullString
}

func getImapMarker(db *sql.DB) (int64, error) {
	var imapMarker int64
	err := db.QueryRow(GET_IMAP_MARKER).Scan(&imapMarker)
	return imapMarker, err
}

func setImapMarker(db *sql.DB, marker int64) (int64, error) {
	var imapMarker int64
	err := db.QueryRow(SET_IMAP_MARKER, marker).Scan(&imapMarker)
	return imapMarker, err
}

func getLatestMessages(imapServer string, username string, password string, appNames map[string]string, db *sql.DB) ([]Prospect, error) {
	//Connect to imap server
	imapClient, err := imap.DialTLS(imapServer, nil)
	if nil != err {
		return nil, err
	}
	defer imapClient.Logout(30 * time.Second)

	log.Println("Server says hello:", imapClient.Data[0].Info)
	imapClient.Data = nil

	//Start TLS
	if imapClient.Caps["STARTTLS"] {
		_, err := imapClient.StartTLS(nil)
		if nil != err {
			return nil, err
		}
	}

	//Authenticate
	if imapClient.State() == imap.Login {
		_, err := imapClient.Login(username, password)
		if nil != err {
			log.Print("Login not successful")
			return nil, err
		} else {
			log.Print("Login successful")
		}
	} else {
		log.Print("Not in login state")
		return nil, err
	}

	// List all top-level mailboxes, wait for the command to finish
	cmd, _ := imap.Wait(imapClient.List("", "%"))

	// Print mailbox information
	log.Println("\nTop-level mailboxes:")
	for _, rsp := range cmd.Data {
		log.Println("|--", rsp.MailboxInfo())
	}

	// Check for new unilateral server data responses
	for _, rsp := range imapClient.Data {
		log.Println("Server data:", rsp)
	}
	imapClient.Data = nil

	//Email regular expression
	log.Print("Compiling e-mail regular expression")
	emailRegex, err := regexp.Compile(EMAIL_REGEX)
	if nil != err {
		return nil, err
	}

	//Open mailbox
	imapClient.Select("INBOX", true)
	log.Print("Mailbox: status\n", imapClient.Mailbox)

	//Prospects array
	var prospects []Prospect

	//Fetch messages
	set, _ := imap.NewSeqSet("")
	latestImapMarket, err := getImapMarker(db)
	if nil != err {
		return nil, err
	}

	if int64(imapClient.Mailbox.Messages) == latestImapMarket {
		log.Print("No new messages")
		return prospects, nil
	} else {
		set.AddRange(uint32(latestImapMarket), imapClient.Mailbox.Messages)
	}

	orderIds := make(map[int64]bool)
	cmd, _ = imapClient.Fetch(set, "RFC822")
	for cmd.InProgress() {
		imapClient.Recv(-1)

		for _, rsp := range cmd.Data {
			if !orderIds[rsp.Order] {
				orderIds[rsp.Order] = true
			} else {
				continue
			}

			msgBytes := imap.AsBytes(rsp.MessageInfo().Attrs["RFC822"])
			if msg, _ := mail.ReadMessage(bytes.NewReader(msgBytes)); nil != msg {
				uA := msg.Header.Get("User-Agent")
				from := msg.Header.Get("From")
				to := msg.Header.Get("To")

				fromEmail := emailRegex.FindString(from)
				toEmail := emailRegex.FindString(to)

				leadId := uuid.NewV3(uuid.Nil, fromEmail)
				appName := appNames[toEmail]

				userAgent := sql.NullString{"", false}
				if len(uA) > 0 {
					userAgent.String = uA
					userAgent.Valid = true
				}

				misc, _ := json.Marshal(msg)
				miscellaneous := sql.NullString{"[" + string(misc) + "]", true}
				prospects = append(prospects, Prospect{leadId, appName, "email", fromEmail, userAgent, miscellaneous})
				body := make([]byte, 20480)
				size, _ := msg.Body.Read(body)
				log.Printf("Read %d bytes\n", size)
			}
		}
	}
	cmd.Data = nil

	//Set IMAP marker
	_, err = setImapMarker(db, int64(imapClient.Mailbox.Messages))

	//Process unilateral server data
	for _, rsp := range imapClient.Data {
		log.Println("Server data:", rsp)
	}
	imapClient.Data = nil

	//Check command completion status
	if rsp, err := cmd.Result(imap.OK); err != nil {
		if err == imap.ErrAborted {
			log.Println("Fetch command aborted")
		} else {
			log.Println("Fetch error:", rsp.Info)
		}
	}

	return prospects, nil
}

func addNewProspects(prospects []Prospect, db *sql.DB) error {
	transaction, err := db.Begin()
	if nil != err {
		log.Print("Error creating transaction")
		log.Print(err)
	}
	defer transaction.Rollback()

	statement, err := transaction.Prepare(QUERY)
	if nil != err {
		log.Print("Error preparing SQL statement")
		log.Print(err)
	}
	defer statement.Close()

	counter := 0
	unused := -1
	for _, prospect := range prospects {
		err = statement.QueryRow(prospect.LeadId.String(), prospect.AppName, prospect.LeadSource, prospect.Email, prospect.UserAgent, prospect.Miscellaneous, time.Now()).Scan(&unused)
		counter++
	}

	err = transaction.Commit()
	if nil != err {
		log.Print("Error committing transaction")
		log.Print(err)
	} else {
		log.Printf("Processed %d prospects", counter)
	}

	return nil
}

func sendEmailReply(smtpServer string, smtpUser string, smtpPassword string, smtpReplyTemplateUrl *url.URL, smtpReplySubject string, prospects []Prospect) error {
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

		smtpClient := gomail.NewDialer(smtpServer, smtpPort, smtpUser, smtpPassword)

		for _, prospect := range prospects {
			message := gomail.NewMessage()
			message.SetHeader("From", smtpUser)
			message.SetHeader("To", prospect.Email)
			message.SetHeader("Subject", smtpReplySubject)
			message.SetBody("text/html", string(smtpReplyTemplate))

			if err := smtpClient.DialAndSend(message); nil != err {
				return err
			}
		}
	}

	return nil
}

func main() {
	//Get app names
	//e.g. APPLICATION_NAMES=info@dipset.com:dipset,info@gunit.com:gunit
	appNames := make(map[string]string)
	appNamesStr := os.Getenv("APPLICATION_NAMES")
	if len(appNamesStr) > 0 {
		appNamesArr := strings.Split(appNamesStr, ",")
		for _, appName := range appNamesArr {
			appNamePair := strings.Split(appName, ":")
			if len(appNamePair) != 2 {
				log.Printf("Invalid application name mapping skipped: %s", appNamePair)
				continue
			}
			appNames[appNamePair[0]] = appNamePair[1]
		}

		if len(appNames) > 0 {
			log.Printf("Application name mappings: %s", appNames)
		} else {
			log.Fatal("No application name mappings set")
		}
	} else {
		log.Fatal("APPLICATION_NAMES variable NOT set")
	}

	//IMAPS server connection
	imapsHost := os.Getenv("IMAPS_HOST")
	imapsUser := os.Getenv("IMAPS_USER")
	imapsPassword := os.Getenv("IMAPS_PASSWORD")

	if len(imapsHost) <= 0 {
		log.Fatal("IMAPS_HOST is NOT set")
	}

	if len(imapsUser) <= 0 {
		log.Fatal("IMAPS_USER is NOT set")
	}

	if len(imapsPassword) <= 0 {
		log.Fatal("IMAPS_PASSWORD is NOT set")
	}

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
		log.Fatalf("SMTP reply template URL is invalid: %s", smtpReplyTemplateUrlStr)
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
	dbCredentials := common.DatabaseCredentials{DB_DRIVER, dbUrl, dbUser, dbPassword, dbName, dbHost, dbPort, dbMaxOpenConns, dbMaxIdleConns}
	if !dbCredentials.IsValid() {
		log.Fatalf("Database credentials NOT set correctly. %#v", dbCredentials)
	}
	db := dbCredentials.GetDatabase()
	defer db.Close()

	//Get latest e-mail messages
	log.Print("Fetching latest e-mail messages")
	prospects, err := getLatestMessages(imapsHost, imapsUser, imapsPassword, appNames, db)
	if nil != err {
		log.Fatal(err)
	}

	//Add prospects from e-mail messages
	err = addNewProspects(prospects, db)
	if nil != err {
		log.Fatal(err)
	}

	//Send thank you reply
	err = sendEmailReply(smtpHost, smtpUser, smtpPassword, smtpReplyTemplateUrl, smtpReplySubject, prospects)
	if nil != err {
		log.Fatal(err)
	}
}