package main

import (
	"bitbucket.org/savewithus/prospects"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/go-martini/martini"
	_ "github.com/lib/pq"
	"github.com/martini-contrib/binding"
	"github.com/martini-contrib/cors"
	"github.com/martini-contrib/gzip"
	"github.com/martini-contrib/secure"
	"log"
	"math/rand"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
)

const (
	QUERY                = "INSERT INTO prospects.leads(lead_id, app_name, email, lead_source, feedback, referrer, page_referrer, first_name, last_name, phone_number, dob, gender, zip_code, language, user_agent, cookies, geolocation, ip_address, miscellaneous, created_at, updated_at) VALUES($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, POINT($17, $18), $19, $20, $21, $22) RETURNING id;"
	ID_QUERY             = "SELECT last_value, increment_by FROM prospects.leads_id_seq"
	LEAD_SOURCE_QUERY    = "SELECT enum_range(NULL::prospects.lead_source) AS lead_sources"
	VERIFY_LEAD_QUERY    = "UPDATE prospects.leads SET is_valid = true WHERE lead_source IN ('landing', 'email', 'phone') AND lead_id = $1 AND (email = $2 OR phone_number = $3)"
	EMAIL_REGEX          = "^[A-Za-z0-9._%-]+@[A-Za-z0-9.-]+[.][A-Za-z]+$"
	UUID_REGEX           = "^[a-z0-9]{8}-[a-z0-9]{4}-[1-5][a-z0-9]{3}-[a-z0-9]{4}-[a-z0-9]{12}$"
	REQUEST_URL          = "/prospects"
	VERIFY_URL           = "/verify"
	ROBOTS_TXT_URL       = "/robots.txt"
	SITEMAP_XML_URL      = "/sitemap.xml"
	FAVICON_ICO_URL      = "/favicon.ico"
	CONTENT_TYPE_HEADER  = "Content-Type"
	LOCATION_HEADER      = "Location"
	CACHE_CONTROL_HEADER = "Cache-Control"
	EXPIRES_HEADER       = "Expires"
	JSON_CONTENT_TYPE    = "application/json"
	XML_CONTENT_TYPE     = "application/xml"
	TEXT_CONTENT_TYPE    = "text/plain"
	HTML_CONTENT_TYPE    = "text/html"
	XFF_HEADER           = "X-Forwarded-For"
)

type Position int

const (
	First Position = 1 << iota
	Last
)

type CreateHandler func(http.ResponseWriter, *http.Request, ProspectForm) (int, string)
type ErrorHandler func(binding.Errors, http.ResponseWriter)
type NotFoundHandler func(http.ResponseWriter, *http.Request) (int, string)

var stringSizeLimit int
var feedbackSizeLimit int
var appNames map[string]bool
var uuidRegex *regexp.Regexp
var emailRegex *regexp.Regexp
var botDetection common.BotDetection
var leadSources map[string]bool
var gzipResponse bool
var gzipCompressionLevel int
var robotsTxtResponse bool
var sitemapXmlResponse bool
var faviconIcoResponse bool
var verifyLeadRedirectUrls map[string]string

type ProspectForm common.Prospect

func (prospect ProspectForm) Validate(errors binding.Errors, req *http.Request) binding.Errors {
	errors = validateSizeLimit(prospect.LeadId, "leadid", stringSizeLimit, errors)
	errors = validateSizeLimit(prospect.AppName, "appname", stringSizeLimit, errors)
	errors = validateSizeLimit(prospect.Referrer, "referrer", stringSizeLimit, errors)
	errors = validateSizeLimit(prospect.PageReferrer, "pagereferrer", stringSizeLimit, errors)
	errors = validateSizeLimit(prospect.FirstName, "firstname", stringSizeLimit, errors)
	errors = validateSizeLimit(prospect.LastName, "lastname", stringSizeLimit, errors)
	errors = validateSizeLimit(prospect.Email, "email", stringSizeLimit, errors)
	errors = validateSizeLimit(prospect.LeadSource, "leadsource", stringSizeLimit, errors)
	errors = validateSizeLimit(prospect.Feedback, "feedback", feedbackSizeLimit, errors)
	errors = validateSizeLimit(prospect.PhoneNumber, "phonenumber", stringSizeLimit, errors)
	errors = validateSizeLimit(prospect.DateOfBirth, "dob", stringSizeLimit, errors)
	errors = validateSizeLimit(prospect.Gender, "gender", stringSizeLimit, errors)
	errors = validateSizeLimit(prospect.ZipCode, "zipcode", stringSizeLimit, errors)
	errors = validateSizeLimit(prospect.Language, "language", stringSizeLimit, errors)
	errors = validateSizeLimit(prospect.UserAgent, "useragent", stringSizeLimit, errors)
	errors = validateSizeLimit(prospect.Cookies, "cookies", stringSizeLimit, errors)
	errors = validateSizeLimit(prospect.IpAddress, "ipaddress", stringSizeLimit, errors)
	errors = validateSizeLimit(prospect.Miscellaneous, "miscellaneous", stringSizeLimit, errors)

	if len(errors) == 0 {
		if len(prospect.AppName) > 0 && appNames != nil && !appNames[prospect.AppName] {
			message := fmt.Sprintf("Invalid appname \"%s\" specified", prospect.AppName)
			errors = addError(errors, []string{"appname"}, binding.TypeError, message)
		}

		if len(prospect.LeadId) > 0 && !uuidRegex.MatchString(prospect.LeadId) {
			message := fmt.Sprintf("Invalid uuid \"%s\" format specified", prospect.LeadId)
			errors = addError(errors, []string{"leadid"}, binding.TypeError, message)
		}

		if !leadSources[prospect.LeadSource] {
			message := fmt.Sprintf("Invalid lead source \"%s\" specified", prospect.LeadSource)
			errors = addError(errors, []string{"leadsource"}, binding.TypeError, message)
		}

		if prospect.LeadSource == "landing" && len(prospect.Email) == 0 && len(prospect.PhoneNumber) == 0 {
			errors = addError(errors, []string{"leadsource", "email", "phonenumber"}, binding.RequiredError, "Email address or Phone number required with landing lead source.")
		}

		if prospect.LeadSource == "email" && len(prospect.Email) == 0 {
			errors = addError(errors, []string{"leadsource", "email"}, binding.RequiredError, "Email address required with email lead source.")
		}

		if prospect.LeadSource == "phone" && len(prospect.PhoneNumber) == 0 {
			errors = addError(errors, []string{"leadsource", "phonenumber"}, binding.RequiredError, "Phone number required with phone lead source.")
		}

		if prospect.LeadSource == "feedback" && len(prospect.Feedback) == 0 {
			errors = addError(errors, []string{"leadsource", "feedback"}, binding.RequiredError, "Feedback required with feedback lead source.")
		}

		IsNotExtended := func(prospect ProspectForm) bool {
			return len(prospect.FirstName) == 0 && len(prospect.LastName) == 0 && len(prospect.Gender) == 0 && len(prospect.DateOfBirth) == 0 && len(prospect.ZipCode) == 0 && len(prospect.Language) == 0 && len(prospect.Miscellaneous) == 0
		}

		if prospect.LeadSource == "extended" && IsNotExtended(prospect) {
			errors = addError(errors, []string{"leadsource", "extended"}, binding.RequiredError, "First name, last name, gender, date of birth, zip code, language and/or miscellaneous is required with extended lead source.")
		}

		if len(prospect.Email) > 0 && !emailRegex.MatchString(prospect.Email) {
			message := fmt.Sprintf("Invalid email \"%s\" format specified", prospect.Email)
			errors = addError(errors, []string{"email"}, binding.TypeError, message)
		}

		if len(prospect.Miscellaneous) > 0 && !common.IsJSON(prospect.Miscellaneous) {
			message := fmt.Sprintf("Invalid format specified for miscellaneous \"%s\"", prospect.Miscellaneous)
			errors = addError(errors, []string{"miscellaneous"}, binding.TypeError, message)
		}

		if len(prospect.DateOfBirth) > 0 {
			dob, err := time.Parse(time.RFC3339, prospect.DateOfBirth)
			var failed bool

			if nil != err {
				failed = true
			} else {
				age := common.GetAge(dob)
				failed = age < 0 || age > 200
			}

			if failed {
				message := fmt.Sprintf("Invalid date of birth \"%s\" specified", prospect.DateOfBirth)
				errors = addError(errors, []string{"dob"}, binding.TypeError, message)
				log.Print(err)
			}
		}

		if len(prospect.Gender) > 0 && (prospect.Gender != "male" && prospect.Gender != "female") {
			message := fmt.Sprintf("Invalid format specified for gender \"%s\", must be male or female", prospect.Gender)
			errors = addError(errors, []string{"gender"}, binding.TypeError, message)
		}

		if prospect.Latitude > 90.0 || prospect.Latitude < -90.0 {
			message := fmt.Sprintf("Invalid latitude \"%f\" specified", prospect.Latitude)
			errors = addError(errors, []string{"latitude"}, binding.TypeError, message)
		}

		if prospect.Longitude > 180.0 || prospect.Longitude < -180.0 {
			message := fmt.Sprintf("Invalid longitude \"%f\" specified", prospect.Longitude)
			errors = addError(errors, []string{"longitude"}, binding.TypeError, message)
		}

		if botDetection.IsBot(req) {
			message := "Go away spambot! We've alerted the authorities"
			errors = addError(errors, []string{"spambot"}, common.BOT_ERROR, message)
		}
	}

	return errors
}

func validateSizeLimit(field string, fieldName string, sizeLimit int, errors binding.Errors) binding.Errors {
	if len(field) > sizeLimit {
		message := fmt.Sprintf("Field %s size %d is too large", fieldName, len(field))
		errors = addError(errors, []string{fieldName}, binding.TypeError, message)
	}
	return errors
}

func addError(errors binding.Errors, fieldNames []string, classification string, message string) binding.Errors {
	errors = append(errors, binding.Error{
		FieldNames:     fieldNames,
		Classification: classification,
		Message:        message,
	})
	return errors
}

func processIpAddressFromAddr(remoteAddr string) string {
	ip, _, err := net.SplitHostPort(remoteAddr)
	if err == nil {
		return ip
	}

	ip2 := net.ParseIP(remoteAddr)
	if ip2 == nil {
		return ""
	}

	return ip2.String()
}

func processIpAddressFromXFF(req *http.Request, position Position) string {
	ipAddresses := strings.Split(req.Header.Get(XFF_HEADER), ",")

	switch position {
	case Last:
		return strings.TrimSpace(ipAddresses[len(ipAddresses)-1])
	case First:
		fallthrough
	default:
		return strings.TrimSpace(ipAddresses[0])
	}
}

var ipAddressLocation string

func processIpAddress(req *http.Request) string {
	var ipAddress string

	switch ipAddressLocation {
	case "xff_first":
		ipAddress = processIpAddressFromXFF(req, First)
		if len(ipAddress) > 0 {
			break
		}
		fallthrough
	case "xff_last":
		ipAddress = processIpAddressFromXFF(req, Last)
		if len(ipAddress) > 0 {
			break
		}
		fallthrough
	case "normal":
		fallthrough
	default:
		ipAddress = processIpAddressFromAddr(req.RemoteAddr)
		break
	}

	return ipAddress
}

var lastValue, incrementBy int64 = -1, -1

func getNextId(db *sql.DB) int64 {
	if -1 == lastValue {
		err := db.QueryRow(ID_QUERY).Scan(&lastValue, &incrementBy)
		if nil != err {
			log.Print(err)
		}
	}

	if -1 != lastValue {
		lastValue += incrementBy
		return lastValue
	} else {
		log.Print("Could not retrieve last sequence number from database.  Returning random value")
		return 7 + rand.Int63n(int64(^uint64(0)>>1)-7)
	}
}

func getLeadSources(db *sql.DB) string {
	var leadSourcesStr string

	err := db.QueryRow(LEAD_SOURCE_QUERY).Scan(&leadSourcesStr)
	if nil != err {
		log.Print(err)
	} else {
		leadSourcesStr = strings.Trim(leadSourcesStr, "{}")
	}

	return leadSourcesStr
}

var asyncRequest bool
var db *sql.DB
var prospectBatchProcessor *common.BatchProcessor

func processProspect(prospectBatch []interface{}, waitGroup *sync.WaitGroup) {
	log.Printf("Starting batch processing of %d prospects", len(prospectBatch))

	defer waitGroup.Done()

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
	for _, prospectInterface := range prospectBatch {
		prospect := prospectInterface.(*ProspectForm)
		_, err = addProspect(db, prospect, statement)
		if nil != err {
			log.Printf("Error processing prospect %#v", prospect)
			log.Print(err)
			continue
		}

		counter++
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

	var err error

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

	dbCredentials := common.DatabaseCredentials{common.DB_DRIVER, dbUrl, dbUser, dbPassword, dbName, dbHost, dbPort, dbMaxOpenConns, dbMaxIdleConns}
	if !dbCredentials.IsValid() {
		log.Fatalf("Database credentials NOT set correctly. %#v", dbCredentials)
	}

	//Seed random number generator
	log.Print("Seeding random number generator")
	rand.Seed(time.Now().UTC().UnixNano())

	//Database connection
	log.Print("Enabling database connectivity")

	db = dbCredentials.GetDatabase()
	defer db.Close()

	//Get configurable string size limits
	stringSizeLimitStr := common.GetenvWithDefault("STRING_SIZE_LIMIT", "500")
	feedbackSizeLimitStr := common.GetenvWithDefault("FEEDBACK_SIZE_LIMIT", "3000")

	stringSizeLimit, err = strconv.Atoi(stringSizeLimitStr)
	if nil != err {
		stringSizeLimit = 500
		log.Printf("Error setting string size limit from value: %s. Default to %d", stringSizeLimitStr, stringSizeLimit)
		log.Print(err)
	}

	feedbackSizeLimit, err = strconv.Atoi(feedbackSizeLimitStr)
	if nil != err {
		feedbackSizeLimit = 10
		log.Printf("Error setting feedback size limit from value: %s. Default to %d", feedbackSizeLimitStr, feedbackSizeLimit)
		log.Print(err)
	}

	//Allowable Application names
	appNamesStr := os.Getenv("APPLICATION_NAMES")
	if len(appNamesStr) > 0 {
		appNames = make(map[string]bool)

		appNamesArr := strings.Split(appNamesStr, ",")
		for _, appName := range appNamesArr {
			appNames[appName] = true
		}

		log.Printf("Allowable application names: %s", appNamesStr)
	} else {
		log.Print("Any application name available")
	}

	//Allowable lead sources
	leadSourcesStr := getLeadSources(db)
	if len(leadSourcesStr) > 0 {
		leadSources = make(map[string]bool)

		leadSourcesArr := strings.Split(leadSourcesStr, ",")
		for _, leadSource := range leadSourcesArr {
			leadSources[leadSource] = true
		}

		log.Printf("Allowable lead sources: %s", leadSourcesStr)
	} else {
		log.Fatal("Unable to retrieve lead sources from database")
	}

	//UUID regular expression
	log.Print("Compiling uuid regular expression")
	uuidRegex, err = regexp.Compile(UUID_REGEX)
	if nil != err {
		log.Print(err)
		log.Fatalf("UUID regex compilation failed for %s", UUID_REGEX)
	}

	//E-mail regular expression
	log.Print("Compiling e-mail regular expression")
	emailRegex, err = regexp.Compile(EMAIL_REGEX)
	if nil != err {
		log.Print(err)
		log.Fatalf("E-mail regex compilation failed for %s", EMAIL_REGEX)
	}

	//Robot detection field
	botDetectionFieldLocationStr := common.GetenvWithDefault("BOTDETECT_FIELDLOCATION", "body")
	botDetectionFieldName := common.GetenvWithDefault("BOTDETECT_FIELDNAME", "spambot")
	botDetectionFieldValue := common.GetenvWithDefault("BOTDETECT_FIELDVALUE", "")
	botDetectionMustMatchStr := common.GetenvWithDefault("BOTDETECT_MUSTMATCH", "true")
	botDetectionPlayCoyStr := common.GetenvWithDefault("BOTDETECT_PLAYCOY", "true")

	var botDetectionFieldLocation common.RequestLocation

	switch botDetectionFieldLocationStr {
	case "header":
		botDetectionFieldLocation = common.Header
		break
	case "body":
		botDetectionFieldLocation = common.Body
		break
	default:
		botDetectionFieldLocation = common.Body
		log.Printf("Error with int input for field %s with value %s.  Defaulting to Body.", "BOTDETECT_FIELDLOCATION", botDetectionFieldLocationStr)
		break
	}

	botDetectionMustMatch, err := strconv.ParseBool(botDetectionMustMatchStr)
	if nil != err {
		botDetectionMustMatch = true
		log.Printf("Error converting boolean input for field %s with value %s. Defaulting to true.", "BOTDETECT_MUSTMATCH", botDetectionMustMatchStr)
		log.Print(err)
	}

	botDetectionPlayCoy, err := strconv.ParseBool(botDetectionPlayCoyStr)
	if nil != err {
		botDetectionPlayCoy = true
		log.Printf("Error converting boolean input for field %s with value %s. Defaulting to true.", "BOTDETECT_PLAYCOY", botDetectionPlayCoyStr)
		log.Print(err)
	}

	botDetection = common.BotDetection{botDetectionFieldLocation, botDetectionFieldName, botDetectionFieldValue, botDetectionMustMatch, botDetectionPlayCoy}

	log.Printf("Creating robot detection with %#v", botDetection)

	//GZIP response compression
	gzipCompressionLevelStr := common.GetenvWithDefault("GZIP_COMPRESSION_LEVEL", "6")
	gzipCompressionLevel, err = strconv.Atoi(gzipCompressionLevelStr)
	if nil != err {
		gzipCompressionLevel = 6
		log.Printf("Error setting gzip compression level from value: %s. Default to %d", gzipCompressionLevelStr, gzipCompressionLevel)
		log.Print(err)
	} else if gzipCompressionLevel < 1 || gzipCompressionLevel > 9 {
		gzipCompressionLevel = 6
		log.Printf("Error setting gzip compression level from value: %d. Default to %d", gzipCompressionLevelStr, gzipCompressionLevel)
	}

	gzipResponseStr := common.GetenvWithDefault("GZIP_RESPONSE", "true")
	gzipResponse, err = strconv.ParseBool(gzipResponseStr)
	if nil != err {
		gzipResponse = true
		log.Printf("Error converting boolean input for field %s with value %s. Defaulting to true.", "GZIP_RESPONSE", gzipResponseStr)
		log.Print(err)
	}
	if gzipResponse {
		log.Printf("Gzip response encoding enabled with level %d", gzipCompressionLevel)
	} else {
		log.Print("Gzip response encoding disabled")
	}

	//IP address location
	ipAddressLocation = common.GetenvWithDefault("IP_ADDRESS_LOCATION", "normal")

	//Asynchronous database writes
	asyncRequest, err = strconv.ParseBool(common.GetenvWithDefault("ASYNC_REQUEST", "false"))
	if nil != err {
		asyncRequest = false
		log.Printf("Error converting input for field ASYNC_REQUEST. Defaulting to false.")
		log.Print(err)
	}

	asyncRequestSizeStr := common.GetenvWithDefault("ASYNC_REQUEST_SIZE", "100000")
	asyncRequestSize, err := strconv.Atoi(asyncRequestSizeStr)
	if nil != err {
		asyncRequestSize = 100000
		log.Printf("Error converting input for field ASYNC_REQUEST_SIZE. Defaulting to 100000.")
		log.Print(err)
	}

	asyncProcessIntervalStr := common.GetenvWithDefault("ASYNC_PROCESS_INTERVAL", "5")
	asyncProcessInterval, err := strconv.Atoi(asyncProcessIntervalStr)
	if nil != err {
		asyncProcessInterval = 5
		log.Printf("Error converting input for field ASYNC_PROCESS_INTERVAL. Defaulting to 5.")
		log.Print(err)
	}

	if asyncRequest {
		prospectBatchProcessor = common.NewBatchProcessor(processProspect, asyncRequestSize, asyncProcessInterval, dbMaxOpenConns)
		prospectBatchProcessor.Start()
		log.Printf("Asynchronous requests enabled. Request queue size set to %d", asyncRequestSize)
		log.Printf("Asynchronous process interval is %d seconds", asyncProcessInterval)
	}

	//robots.txt
	robotsTxtResponseStr := common.GetenvWithDefault("ROBOTS_TXT", "false")
	robotsTxtResponse, err = strconv.ParseBool(robotsTxtResponseStr)
	if nil != err {
		robotsTxtResponse = false
		log.Printf("Error converting boolean input for field %s with value %s. Defaulting to false.", "ROBOTS_TXT", robotsTxtResponseStr)
		log.Print(err)
	}

	if robotsTxtResponse {
		log.Print("robots.txt support enabled")
	} else {
		log.Print("robots.txt support disabled")
	}

	//sitemap.xml
	sitemapXmlResponseStr := common.GetenvWithDefault("SITEMAP_XML", "false")
	sitemapXmlResponse, err = strconv.ParseBool(sitemapXmlResponseStr)
	if nil != err {
		sitemapXmlResponse = false
		log.Printf("Error converting boolean input for field %s with value %s. Defaulting to false.", "SITEMAP_XML", sitemapXmlResponseStr)
		log.Print(err)
	}

	if sitemapXmlResponse {
		log.Print("sitemap.xml support enabled")
	} else {
		log.Print("sitemap.xml support disabled")
	}

	//favicon.ico
	faviconIcoResponseStr := common.GetenvWithDefault("FAVICON_ICO", "false")
	faviconIcoResponse, err = strconv.ParseBool(faviconIcoResponseStr)
	if nil != err {
		faviconIcoResponse = false
		log.Printf("Error converting boolean input for field %s with value %s. Defaulting to false.", "FAVICON_ICO", faviconIcoResponseStr)
		log.Print(err)
	}

	if faviconIcoResponse {
		log.Print("favicon.ico support enabled")
	} else {
		log.Print("favicon.ico support disabled")
	}

	//Verify leads
	verifyLeadRedirectUrlsStr := os.Getenv("VERIFY_LEAD_REDIRECT_URLS")
	if len(verifyLeadRedirectUrlsStr) > 0 {
		verifyLeadRedirectUrls = make(map[string]string)

		isUrl := func(val string) bool {
			_, err := url.Parse(val)
			return nil == err
		}

		verifyLeadRedirectUrlsArr := strings.Split(verifyLeadRedirectUrlsStr, ",")
		for _, verifyLeadRedirectUrl := range verifyLeadRedirectUrlsArr {
			nvp := strings.Split(verifyLeadRedirectUrl, "|")

			if len(nvp) == 2 && isUrl(nvp[1]) {
				verifyLeadRedirectUrls[nvp[0]] = nvp[1]
				log.Printf("Added lead verification redirect url: %s for app name %s", nvp[1], nvp[0])
			} else {
				log.Printf("Invalid lead verification redirect url specified %s", verifyLeadRedirectUrl)
			}
		}
	} else {
		log.Print("No lead verification redirect urls configured")
	}

	//Signal handler
	signals := make(chan os.Signal)
	signal.Notify(signals, os.Interrupt)
	signal.Notify(signals, syscall.SIGTERM)
	go func() {
		<-signals
		log.Print("Shutting down...")
		if nil != prospectBatchProcessor {
			prospectBatchProcessor.Stop()
			log.Print("Prospect batch processor shut down")
		}
		os.Exit(0)
	}()

	//HTTP handlers
	log.Print("Preparing HTTP handlers")
	createHandler, errorHandler, notFoundHandler := setupHttpHandlers(db)

	//HTTP server
	host := common.GetenvWithDefault("HOST", "")
	port := common.GetenvWithDefault("PORT", "3000")
	mode := common.GetenvWithDefault("MARTINI_ENV", "development")

	log.Printf("Running HTTP server on %s:%s in mode %s", host, port, mode)
	runHttpServer(createHandler, errorHandler, notFoundHandler)
}

func setupHttpHandlers(db *sql.DB) (CreateHandler, ErrorHandler, NotFoundHandler) {
	createHandler := func(res http.ResponseWriter, req *http.Request, prospect ProspectForm) (int, string) {
		prospect.IpAddress = processIpAddress(req)
		prospect.Referrer = req.Referer()
		prospect.UserAgent = req.UserAgent()

		var cookiesArr []string
		for _, cookie := range req.Cookies() {
			cookiesArr = append(cookiesArr, cookie.String())
		}
		prospect.Cookies = strings.Join(cookiesArr, ", ")

		if len(prospect.Cookies) > 0 {
			prospect.Cookies = fmt.Sprintf("{%s}", prospect.Cookies)
		}

		log.Printf("Received new prospect: %#v", prospect)

		req.Close = true
		res.Header().Set(CONTENT_TYPE_HEADER, JSON_CONTENT_TYPE)
		var response common.Response

		if asyncRequest && prospectBatchProcessor.Running {
			prospectBatchProcessor.AddEvent(&prospect)
			responseStr := "Successfully added prospect"
			response = common.Response{Code: http.StatusAccepted, Message: responseStr}
			log.Print(responseStr)
		} else if asyncRequest && !prospectBatchProcessor.Running {
			responseStr := "Could not add prospect due to server maintenance"
			response = common.Response{Code: http.StatusServiceUnavailable, Message: responseStr}
			log.Print(responseStr)
		} else {
			id, err := addProspect(db, &prospect, nil)
			if nil != err {
				responseStr := "Could not add prospect due to server error"
				response = common.Response{Code: http.StatusInternalServerError, Message: responseStr}
				log.Print(responseStr)
				log.Print(err)
				log.Printf("%d database connections opened", db.Stats().OpenConnections)
			} else {
				responseStr := "Successfully added prospect"
				response = common.Response{Code: http.StatusCreated, Message: responseStr, Id: id}
				log.Print(responseStr)
			}
		}

		jsonStr, _ := json.Marshal(response)
		return response.Code, string(jsonStr)
	}

	errorHandler := func(errors binding.Errors, res http.ResponseWriter) {
		if len(errors) > 0 {
			var fieldsMsg string

			for _, err := range errors {
				for _, field := range err.Fields() {
					fieldsMsg += fmt.Sprintf("%s, ", field)
				}

				log.Printf("Error received. Message: %s, Kind: %s", err.Error(), err.Kind())
			}

			fieldsMsg = strings.TrimSuffix(fieldsMsg, ", ")

			log.Printf("Error received. Fields: %s", fieldsMsg)

			res.Header().Set(CONTENT_TYPE_HEADER, JSON_CONTENT_TYPE)
			var response common.Response

			if errors.Has(binding.RequiredError) {
				res.WriteHeader(http.StatusBadRequest)
				responseStr := fmt.Sprintf("Missing required field(s): %s", fieldsMsg)
				response = common.Response{Code: http.StatusBadRequest, Message: responseStr}
			} else if errors.Has(binding.ContentTypeError) {
				res.WriteHeader(http.StatusUnsupportedMediaType)
				response = common.Response{Code: http.StatusUnsupportedMediaType, Message: "Invalid content type"}
			} else if errors.Has(binding.DeserializationError) {
				res.WriteHeader(http.StatusBadRequest)
				response = common.Response{Code: http.StatusBadRequest, Message: "Deserialization error"}
			} else if errors.Has(binding.TypeError) {
				res.WriteHeader(http.StatusBadRequest)
				response = common.Response{Code: http.StatusBadRequest, Message: errors[0].Error()}
			} else if errors.Has(common.BOT_ERROR) {
				if botDetection.PlayCoy && !asyncRequest {
					res.WriteHeader(http.StatusCreated)
					response = common.Response{Code: http.StatusCreated, Message: "Successfully added prospect", Id: getNextId(db)}
					log.Printf("Robot detected: %s. Playing coy.", errors[0].Error())
				} else if botDetection.PlayCoy && asyncRequest {
					res.WriteHeader(http.StatusAccepted)
					response = common.Response{Code: http.StatusAccepted, Message: "Successfully added prospect"}
					log.Printf("Robot detected: %s. Playing coy.", errors[0].Error())
				} else {
					res.WriteHeader(http.StatusBadRequest)
					response = common.Response{Code: http.StatusBadRequest, Message: errors[0].Error()}
					log.Printf("Robot detected: %s. Rejecting message.", errors[0].Error())
				}
			} else {
				res.WriteHeader(http.StatusBadRequest)
				response = common.Response{Code: http.StatusBadRequest, Message: "Unknown error"}
			}

			log.Print(response.Message)
			jsonStr, _ := json.Marshal(response)
			res.Write(jsonStr)
		}
	}

	notFoundHandler := func(res http.ResponseWriter, req *http.Request) (int, string) {
		req.Close = true
		res.Header().Set(CONTENT_TYPE_HEADER, JSON_CONTENT_TYPE)
		responseStr := fmt.Sprintf("URL Not Found %s", req.URL)
		response := common.Response{Code: http.StatusNotFound, Message: responseStr}
		log.Print(responseStr)
		jsonStr, _ := json.Marshal(response)
		return response.Code, string(jsonStr)
	}

	return createHandler, errorHandler, notFoundHandler
}

func addProspect(db *sql.DB, prospect *ProspectForm, statement *sql.Stmt) (int64, error) {
	var email sql.NullString
	if len(prospect.Email) != 0 {
		email = sql.NullString{prospect.Email, true}
	}

	var feedback sql.NullString
	if len(prospect.Feedback) != 0 {
		feedback = sql.NullString{prospect.Feedback, true}
	}

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

	var pageReferrer sql.NullString
	if len(prospect.PageReferrer) != 0 {
		pageReferrer = sql.NullString{prospect.PageReferrer, true}
	}

	var dob sql.NullString
	if len(prospect.DateOfBirth) != 0 {
		dob = sql.NullString{prospect.DateOfBirth, true}
	}

	var gender sql.NullString
	if len(prospect.Gender) != 0 {
		gender = sql.NullString{prospect.Gender, true}
	}

	var zipCode sql.NullString
	if len(prospect.ZipCode) != 0 {
		zipCode = sql.NullString{prospect.ZipCode, true}
	}

	var language sql.NullString
	if len(prospect.Language) != 0 {
		language = sql.NullString{prospect.Language, true}
	}

	var userAgent sql.NullString
	if len(prospect.UserAgent) != 0 {
		userAgent = sql.NullString{prospect.UserAgent, true}
	}

	var latitude sql.NullFloat64
	var longitude sql.NullFloat64
	if prospect.Latitude != 0 && prospect.Longitude != 0 {
		latitude = sql.NullFloat64{prospect.Latitude, true}
		longitude = sql.NullFloat64{prospect.Longitude, true}
	}

	var ipAddress sql.NullString
	if len(prospect.IpAddress) != 0 {
		ipAddress = sql.NullString{prospect.IpAddress, true}
	}

	var miscellaneous sql.NullString
	if len(prospect.Miscellaneous) != 0 {
		miscellaneous = sql.NullString{prospect.Miscellaneous, true}
	}

	var cookies sql.NullString
	if len(prospect.Cookies) != 0 {
		cookies = sql.NullString{prospect.Cookies, true}
	}

	var lastInsertId int64
	var err error
	if nil == statement {
		err = db.QueryRow(QUERY, prospect.LeadId, prospect.AppName, email, prospect.LeadSource, feedback, referrer, pageReferrer, firstName, lastName, phoneNumber, dob, gender, zipCode, language, userAgent, cookies, latitude, longitude, ipAddress, miscellaneous, time.Now(), time.Now()).Scan(&lastInsertId)
	} else {
		err = statement.QueryRow(prospect.LeadId, prospect.AppName, email, prospect.LeadSource, feedback, referrer, pageReferrer, firstName, lastName, phoneNumber, dob, gender, zipCode, language, userAgent, cookies, latitude, longitude, ipAddress, miscellaneous, time.Now(), time.Now()).Scan(&lastInsertId)
	}

	if nil == err {
		log.Printf("New prospect id = %d", lastInsertId)
	}

	return lastInsertId, err
}

func runHttpServer(createHandler CreateHandler, errorHandler ErrorHandler, notFoundHandler NotFoundHandler) {
	martini_ := martini.Classic()

	allowHeaders := []string{"Origin"}
	if botDetection.FieldLocation == common.Header {
		allowHeaders = append(allowHeaders, botDetection.FieldName)
	}

	//Allowable header names
	headerNamesStr := os.Getenv("ALLOW_HEADERS")
	if len(headerNamesStr) > 0 {
		headerNamesArr := strings.Split(headerNamesStr, ",")
		for _, headerName := range headerNamesArr {
			allowHeaders = append(allowHeaders, headerName)
		}
	}

	log.Printf("Allowable header names: %s", allowHeaders)

	//GZIP responses
	if gzipResponse {
		martini_.Use(gzip.All(gzip.Options{CompressionLevel: gzipCompressionLevel}))
	}

	martini_.Use(cors.Allow(&cors.Options{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{common.POST_METHOD, common.GET_METHOD, common.HEAD_METHOD},
		AllowHeaders:     allowHeaders,
		AllowCredentials: true,
	}))

	sslRedirect, err := strconv.ParseBool(common.GetenvWithDefault("SSL_REDIRECT", "false"))
	if nil != err {
		sslRedirect = false
		log.Print(err)
	}
	log.Printf("Setting SSL redirect to %t", sslRedirect)

	martini_.Use(secure.Secure(secure.Options{
		SSLRedirect: sslRedirect,
	}))

	//robots.txt
	if robotsTxtResponse {
		getRobotsTxt := func(res http.ResponseWriter, req *http.Request) (int, string) {
			res.Header().Set(CONTENT_TYPE_HEADER, TEXT_CONTENT_TYPE)
			var robotsTxt common.RobotsTxt
			robotsTxt.AddRecord(common.RobotsRecord{[]string{"*"}, []string{"/"}})
			return http.StatusOK, robotsTxt.String()
		}
		martini_.Get(ROBOTS_TXT_URL, getRobotsTxt, errorHandler)
		martini_.Head(ROBOTS_TXT_URL, getRobotsTxt, errorHandler)
	}

	//sitemap.xml
	if sitemapXmlResponse {
		getSitemapXml := func(res http.ResponseWriter, req *http.Request) (int, string) {
			res.Header().Set(CONTENT_TYPE_HEADER, XML_CONTENT_TYPE)

			hostname := fmt.Sprintf("%s://%s", common.GetScheme(req), req.Host)
			var urlSet common.UrlSet

			url := fmt.Sprintf("%s%s", hostname, REQUEST_URL)
			urlSet.AddUrl(common.Url{Location: url, LastModification: time.Now(), ChangeFrequency: common.Always, Priority: 1.0})
			return http.StatusOK, urlSet.String()
		}
		martini_.Get(SITEMAP_XML_URL, getSitemapXml, errorHandler)
		martini_.Head(SITEMAP_XML_URL, getSitemapXml, errorHandler)
	}

	//favicon.ico
	if faviconIcoResponse {
		getFaviconIco := func(res http.ResponseWriter, req *http.Request) (int, string) {
			res.Header().Set(CONTENT_TYPE_HEADER, common.ICO_CONTENT_TYPE)
			var favicon common.FaviconIco
			favicon.CreateImage(0xA384C1FF, 16, 16)
			return http.StatusOK, string(favicon.GetImageData())
		}
		martini_.Get(FAVICON_ICO_URL, getFaviconIco, errorHandler)
		martini_.Head(FAVICON_ICO_URL, getFaviconIco, errorHandler)
	}

	//Verify prospect
	if len(verifyLeadRedirectUrls) > 0 {
		notFoundText := `<!DOCTYPE html><html><head><meta charset="UTF-8"><title>Not found!</title></head>
                         <body><img src="http://media02.hongkiat.com/error_404_01/csstricks.jpg" alt="Not found css"></body></html>`

		getSqlString := func(values *url.Values, fieldName string) sql.NullString {
			value := values.Get(fieldName)
			empty := len(value) <= 0
			return sql.NullString{value, !empty}
		}

		isValidUserId := func(val *sql.NullString) bool {
			return val.Valid && uuidRegex.MatchString(val.String)
		}

		isValidEmail := func(val *sql.NullString) bool {
			return val.Valid && emailRegex.MatchString(val.String)
		}

		verifyProspect := func(res http.ResponseWriter, req *http.Request) (int, string) {
			go func() {
				values := req.URL.Query()
				userId := getSqlString(&values, "userId")
				email := getSqlString(&values, "email")
				phoneNumber := getSqlString(&values, "phone_number")

				if isValidUserId(&userId) && (isValidEmail(&email) || phoneNumber.Valid) {
					res, err := db.Exec(VERIFY_LEAD_QUERY, userId, email, phoneNumber)
					if nil != err {
						log.Print(err)
					} else {
						count, _ := res.RowsAffected()
						if count == 0 {
							log.Printf("No leads verified after request %s", values)
						} else {
							log.Printf("Successfully verified %d leads after request %s", count, values)
						}
					}
				} else {
					log.Printf("Invalid verification request received: %s", values)
				}
			}()

			appName := req.URL.Query().Get("app_name")
			verifyRedirectUrl := verifyLeadRedirectUrls[appName]

			if len(verifyRedirectUrl) == 0 {
				res.Header().Set(CONTENT_TYPE_HEADER, HTML_CONTENT_TYPE)
				return http.StatusNotFound, notFoundText
			} else {

				verifyResponseText := fmt.Sprintf(`<!DOCTYPE html><html><head><meta charset="UTF-8"><title>Redirect in progress</title></head>
                                                   <body><p>Please visit us <a href="%s">here</a></p></body></html>`, verifyRedirectUrl)

				res.Header().Set(LOCATION_HEADER, verifyRedirectUrl)
				res.Header().Set(CONTENT_TYPE_HEADER, HTML_CONTENT_TYPE)
				res.Header().Set(CACHE_CONTROL_HEADER, "no-store, no-cache, must-revalidate")
				res.Header().Set(EXPIRES_HEADER, "Thu, 01 Jan 1970 00:00:00 GMT")
				return http.StatusMovedPermanently, verifyResponseText
			}
		}
		martini_.Get(VERIFY_URL, verifyProspect, errorHandler)
		martini_.Head(VERIFY_URL, verifyProspect, errorHandler)
	}

	//Prospects
	martini_.Post(REQUEST_URL, binding.Form(ProspectForm{}), errorHandler, createHandler)
	martini_.NotFound(notFoundHandler)

	//Event loop
	martini_.Run()
}
