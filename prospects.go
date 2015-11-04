package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/go-martini/martini"
	_ "github.com/lib/pq"
	"github.com/martini-contrib/binding"
	"github.com/martini-contrib/cors"
	"github.com/martini-contrib/secure"
	"log"
	"math/rand"
	"net"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const (
	QUERY               = "INSERT INTO prospects.leads(lead_id, app_name, email, used_pinterest, used_facebook, used_instagram, used_twitter, used_google, used_youtube, extended, feedback, referrer, page_referrer, first_name, last_name, phone_number, dob, gender, zip_code, language, user_agent, cookies, geolocation, ip_address, miscellaneous, created_at) VALUES($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, POINT($23, $24), $25, $26, $27) RETURNING id;"
	ID_QUERY            = "SELECT last_value, increment_by FROM prospects.leads_id_seq"
	EMAIL_REGEX         = "^[A-Za-z0-9._%-]+@[A-Za-z0-9.-]+[.][A-Za-z]+$"
	UUID_REGEX          = "^[a-z0-9]{8}-[a-z0-9]{4}-[1-5][a-z0-9]{3}-[a-z0-9]{4}-[a-z0-9]{12}$"
	POST_URL            = "/prospects"
	DB_DRIVER           = "postgres"
	CONTENT_TYPE_NAME   = "Content-Type"
	JSON_CONTENT_TYPE   = "application/json"
	STRING_SIZE_LIMIT   = 500
	FEEDBACK_SIZE_LIMIT = 3000
	BotError            = "BotError"
)

type ProspectForm struct {
	LeadId        string `form:"leadid" binding:"required"`
	AppName       string `form:"appname" binding:"required"`
	Referrer      string
	PageReferrer  string `form:"pagereferrer"`
	FirstName     string `form:"firstname"`
	LastName      string `form:"lastname"`
	Email         string `form:"email"`
	Pinterest     bool   `form:"pinterest"`
	Facebook      bool   `form:"facebook"`
	Instagram     bool   `form:"instagram"`
	Twitter       bool   `form:"twitter"`
	Google        bool   `form:"google"`
	Youtube       bool   `form:"youtube"`
	Extended      bool   `form:"extended"`
	Feedback      string `form:"feedback"`
	PhoneNumber   string `form:"phonenumber"`
	DateOfBirth   string `form:"dob"`
	Gender        string `form:"gender"`
	ZipCode       string `form:"zipcode"`
	Language      string `form:"language"`
	UserAgent     string
	Cookies       string
	Latitude      float64 `form:"latitude"`
	Longitude     float64 `form:"longitude"`
	IpAddress     string
	Miscellaneous string `form:"miscellaneous"`
}

type Response struct {
	Code    int
	Message string
	Id      int64 `json:",omitempty"`
}

type RequestLocation int

const (
	Header RequestLocation = 1 << iota
	Body
)

type BotDetection struct {
	FieldLocation RequestLocation
	FieldName     string
	FieldValue    string
	MustMatch     bool
	PlayCoy       bool
}

func (botDetection BotDetection) IsBot(req *http.Request) bool {
	var botField string

	switch botDetection.FieldLocation {
	case Header:
		botField = req.Header.Get(botDetection.FieldName)
		break
	case Body:
		botField = req.FormValue(botDetection.FieldName)
		break
	}

	if botDetection.MustMatch && botDetection.FieldValue == botField {
		return false
	} else if !botDetection.MustMatch && botDetection.FieldValue != botField {
		return false
	}

	return true
}

type CreateHandler func(http.ResponseWriter, *http.Request, ProspectForm) (int, string)
type ErrorHandler func(binding.Errors, http.ResponseWriter)
type NotFoundHandler func(http.ResponseWriter, *http.Request) (int, string)

var appNames map[string]bool
var uuidRegex *regexp.Regexp
var emailRegex *regexp.Regexp
var botDetection BotDetection

func (prospect ProspectForm) Validate(errors binding.Errors, req *http.Request) binding.Errors {
	errors = validateSizeLimit(prospect.LeadId, "leadid", STRING_SIZE_LIMIT, errors)
	errors = validateSizeLimit(prospect.AppName, "appname", STRING_SIZE_LIMIT, errors)
	errors = validateSizeLimit(prospect.Referrer, "referrer", STRING_SIZE_LIMIT, errors)
	errors = validateSizeLimit(prospect.PageReferrer, "pagereferrer", STRING_SIZE_LIMIT, errors)
	errors = validateSizeLimit(prospect.FirstName, "firstname", STRING_SIZE_LIMIT, errors)
	errors = validateSizeLimit(prospect.LastName, "lastname", STRING_SIZE_LIMIT, errors)
	errors = validateSizeLimit(prospect.Email, "email", STRING_SIZE_LIMIT, errors)
	errors = validateSizeLimit(prospect.Feedback, "feedback", FEEDBACK_SIZE_LIMIT, errors)
	errors = validateSizeLimit(prospect.PhoneNumber, "phonenumber", STRING_SIZE_LIMIT, errors)
	errors = validateSizeLimit(prospect.DateOfBirth, "dob", STRING_SIZE_LIMIT, errors)
	errors = validateSizeLimit(prospect.Gender, "gender", STRING_SIZE_LIMIT, errors)
	errors = validateSizeLimit(prospect.ZipCode, "zipcode", STRING_SIZE_LIMIT, errors)
	errors = validateSizeLimit(prospect.Language, "language", STRING_SIZE_LIMIT, errors)
	errors = validateSizeLimit(prospect.UserAgent, "useragent", STRING_SIZE_LIMIT, errors)
	errors = validateSizeLimit(prospect.Cookies, "cookies", STRING_SIZE_LIMIT, errors)
	errors = validateSizeLimit(prospect.IpAddress, "ipaddress", STRING_SIZE_LIMIT, errors)
	errors = validateSizeLimit(prospect.Miscellaneous, "miscellaneous", STRING_SIZE_LIMIT, errors)

	if len(errors) == 0 {
		if len(prospect.AppName) > 0 && appNames != nil && !appNames[prospect.AppName] {
			message := fmt.Sprintf("Invalid appname \"%s\" specified", prospect.AppName)
			errors = addError(errors, []string{"appname"}, binding.TypeError, message)
		}

		if len(prospect.LeadId) > 0 && !uuidRegex.MatchString(prospect.LeadId) {
			message := fmt.Sprintf("Invalid uuid \"%s\" format specified", prospect.LeadId)
			errors = addError(errors, []string{"leadid"}, binding.TypeError, message)
		}

		invalidId := len(prospect.Email) == 0 && len(prospect.PhoneNumber) == 0 && !prospect.Pinterest && !prospect.Facebook && !prospect.Instagram && !prospect.Twitter && !prospect.Google && !prospect.Youtube && !prospect.Extended && len(prospect.Feedback) == 0
		if invalidId {
			errors = addError(errors, []string{"email", "phonenumber", "pinterest", "facebook", "instagram", "twitter", "google", "youtube", "extended", "feedback"}, binding.RequiredError, "At least one of email, pinterest, facebook, instagram, twitter, google, youtube, extended, or feedback is required")
		}

		if len(prospect.Email) > 0 && !emailRegex.MatchString(prospect.Email) {
			message := fmt.Sprintf("Invalid email \"%s\" format specified", prospect.Email)
			errors = addError(errors, []string{"email"}, binding.TypeError, message)
		}

		if len(prospect.Miscellaneous) > 0 && !isJSON(prospect.Miscellaneous) {
			message := fmt.Sprintf("Invalid format specified for miscellaneous \"%s\"", prospect.Miscellaneous)
			errors = addError(errors, []string{"miscellaneous"}, binding.TypeError, message)
		}

		if len(prospect.DateOfBirth) > 0 {
			dob, err := time.Parse(time.RFC3339, prospect.DateOfBirth)
			var failed bool

			if nil != err {
				failed = true
			} else {
				age := getAge(dob)
				failed = age < 0 || age > 200
			}

			if failed {
				message := fmt.Sprintf("Invalid date of birth \"%s\" specified", prospect.DateOfBirth)
				errors = addError(errors, []string{"dob"}, binding.TypeError, message)
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
			errors = addError(errors, []string{"spambot"}, BotError, message)
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

func GetenvWithDefault(envKey string, defaultVal string) string {
	envVal := os.Getenv(envKey)

	if len(envVal) == 0 {
		envVal = defaultVal
	}

	return envVal
}

func processIpAddress(remoteAddr string) string {
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

func isJSON(str string) bool {
	var js map[string]interface{}
	return json.Unmarshal([]byte(str), &js) == nil
}

func getAge(dob time.Time) int64 {
	age := time.Now().Sub(dob).Seconds() / 31536000
	return int64(age)
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

func main() {
	//Seed random number generator
	rand.Seed(time.Now().UTC().UnixNano())

	//Database connection
	log.Print("Enabling database connectivity")

	dbUrl := os.Getenv("DATABASE_URL")
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")
	dbHost := GetenvWithDefault("DB_HOST", "localhost")
	dbPort := GetenvWithDefault("DB_PORT", "5432")
	dbMaxOpenConns := GetenvWithDefault("DB_MAX_OPEN_CONNS", "10")
	dbMaxIdleConns := GetenvWithDefault("DB_MAX_IDLE_CONNS", "0")

	dbCredentials := DatabaseCredentials{DB_DRIVER, dbUrl, dbUser, dbPassword, dbName, dbHost, dbPort, dbMaxOpenConns, dbMaxIdleConns}
	if !dbCredentials.IsValid() {
		log.Fatalf("Database credentials NOT set correctly. %#v", dbCredentials)
	}

	db := dbCredentials.GetDatabase()
	defer db.Close()

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

	var err error

	//UUID regular expression
	log.Print("Compiling uuid regular expression")
	uuidRegex, err = regexp.Compile(UUID_REGEX)
	if nil != err {
		log.Fatalf("UUID regex compilation failed for %s", UUID_REGEX)
	}

	//E-mail regular expression
	log.Print("Compiling e-mail regular expression")
	emailRegex, err = regexp.Compile(EMAIL_REGEX)
	if nil != err {
		log.Fatalf("E-mail regex compilation failed for %s", EMAIL_REGEX)
	}

	//Robot detection field
	botDetectionFieldLocationStr := GetenvWithDefault("BOTDETECT_FIELDLOCATION", "body")
	botDetectionFieldName := GetenvWithDefault("BOTDETECT_FIELDNAME", "spambot")
	botDetectionFieldValue := GetenvWithDefault("BOTDETECT_FIELDVALUE", "")
	botDetectionMustMatchStr := GetenvWithDefault("BOTDETECT_MUSTMATCH", "true")
	botDetectionPlayCoyStr := GetenvWithDefault("BOTDETECT_PLAYCOY", "true")

	var botDetectionFieldLocation RequestLocation

	switch botDetectionFieldLocationStr {
	case "header":
		botDetectionFieldLocation = Header
		break
	case "body":
		botDetectionFieldLocation = Body
		break
	default:
		botDetectionFieldLocation = Body
		log.Printf("Error with int input for field %s with value %s.  Defaulting to Body.", "BOTDETECT_FIELDLOCATION", botDetectionFieldLocationStr)
		break
	}

	botDetectionMustMatch, err := strconv.ParseBool(botDetectionMustMatchStr)
	if nil != err {
		botDetectionMustMatch = true
		log.Printf("Error converting boolean input for field %s with value %s. Defaulting to true.", "BOTDETECT_MUSTMATCH", botDetectionMustMatchStr)
	}

	botDetectionPlayCoy, err := strconv.ParseBool(botDetectionPlayCoyStr)
	if nil != err {
		botDetectionPlayCoy = true
		log.Printf("Error converting boolean input for field %s with value %s. Defaulting to true.", "BOTDETECT_PLAYCOY", botDetectionPlayCoyStr)
	}

	botDetection = BotDetection{botDetectionFieldLocation, botDetectionFieldName, botDetectionFieldValue, botDetectionMustMatch, botDetectionPlayCoy}

	log.Printf("Creating robot detection with %#v", botDetection)

	//HTTP handlers
	log.Print("Preparing HTTP handlers")
	createHandler, errorHandler, notFoundHandler := setupHttpHandlers(db)

	//HTTP server
	host := GetenvWithDefault("HOST", "")
	port := GetenvWithDefault("PORT", "3000")
	mode := GetenvWithDefault("MARTINI_ENV", "development")

	log.Printf("Running HTTP server on %s:%s in mode %s", host, port, mode)
	runHttpServer(createHandler, errorHandler, notFoundHandler)
}

func setupHttpHandlers(db *sql.DB) (CreateHandler, ErrorHandler, NotFoundHandler) {
	createHandler := func(res http.ResponseWriter, req *http.Request, prospect ProspectForm) (int, string) {
		prospect.IpAddress = processIpAddress(req.RemoteAddr)
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
		res.Header().Set(CONTENT_TYPE_NAME, JSON_CONTENT_TYPE)
		var response Response

		id, err := addProspect(db, prospect)
		if nil != err {
			responseStr := fmt.Sprintf("Could not add prospect due to server error: e-mail %s", prospect.Email)
			response = Response{Code: http.StatusInternalServerError, Message: responseStr}
			log.Print(responseStr)
			log.Print(err)
			log.Printf("%d database connections opened", db.Stats().OpenConnections)
		} else {
			responseStr := "Successfully added prospect"
			response = Response{Code: http.StatusCreated, Message: responseStr, Id: id}
			log.Print(responseStr)
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

			res.Header().Set(CONTENT_TYPE_NAME, JSON_CONTENT_TYPE)
			var response Response

			if errors.Has(binding.RequiredError) {
				res.WriteHeader(http.StatusBadRequest)
				responseStr := fmt.Sprintf("Missing required field(s): %s", fieldsMsg)
				response = Response{Code: http.StatusBadRequest, Message: responseStr}
			} else if errors.Has(binding.ContentTypeError) {
				res.WriteHeader(http.StatusUnsupportedMediaType)
				response = Response{Code: http.StatusUnsupportedMediaType, Message: "Invalid content type"}
			} else if errors.Has(binding.DeserializationError) {
				res.WriteHeader(http.StatusBadRequest)
				response = Response{Code: http.StatusBadRequest, Message: "Deserialization error"}
			} else if errors.Has(binding.TypeError) {
				res.WriteHeader(http.StatusBadRequest)
				response = Response{Code: http.StatusBadRequest, Message: errors[0].Error()}
			} else if errors.Has(BotError) {
				if botDetection.PlayCoy {
					res.WriteHeader(http.StatusCreated)
					response = Response{Code: http.StatusCreated, Message: "Successfully added prospect", Id: getNextId(db)}
					log.Printf("Robot detected: %s. Playing coy.", errors[0].Error())
				} else {
					res.WriteHeader(http.StatusBadRequest)
					response = Response{Code: http.StatusBadRequest, Message: errors[0].Error()}
					log.Printf("Robot detected: %s. Rejecting message.", errors[0].Error())
				}
			} else {
				res.WriteHeader(http.StatusBadRequest)
				response = Response{Code: http.StatusBadRequest, Message: "Unknown error"}
			}

			log.Print(response.Message)
			jsonStr, _ := json.Marshal(response)
			res.Write(jsonStr)
		}
	}

	notFoundHandler := func(res http.ResponseWriter, req *http.Request) (int, string) {
		req.Close = true
		res.Header().Set(CONTENT_TYPE_NAME, JSON_CONTENT_TYPE)
		responseStr := fmt.Sprintf("URL Not Found %s", req.URL)
		response := Response{Code: http.StatusNotFound, Message: responseStr}
		log.Print(responseStr)
		jsonStr, _ := json.Marshal(response)
		return response.Code, string(jsonStr)
	}

	return createHandler, errorHandler, notFoundHandler
}

func addProspect(db *sql.DB, prospect ProspectForm) (int64, error) {
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
	err := db.QueryRow(QUERY, prospect.LeadId, prospect.AppName, email, prospect.Pinterest, prospect.Facebook, prospect.Instagram, prospect.Twitter, prospect.Google, prospect.Youtube, prospect.Extended, feedback, referrer, pageReferrer, firstName, lastName, phoneNumber, dob, gender, zipCode, language, userAgent, cookies, latitude, longitude, ipAddress, miscellaneous, time.Now()).Scan(&lastInsertId)

	if nil == err {
		log.Printf("New prospect id = %d", lastInsertId)
	}

	return lastInsertId, err
}

func runHttpServer(createHandler CreateHandler, errorHandler ErrorHandler, notFoundHandler NotFoundHandler) {
	martini_ := martini.Classic()

	allowHeaders := []string{"Origin"}
	if botDetection.FieldLocation == Header {
		allowHeaders = append(allowHeaders, botDetection.FieldName)
	}

	martini_.Use(cors.Allow(&cors.Options{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"POST"},
		AllowHeaders:     allowHeaders,
		AllowCredentials: true,
	}))

	sslRedirect, err := strconv.ParseBool(GetenvWithDefault("SSL_REDIRECT", "false"))
	if nil != err {
		sslRedirect = false
	}

	martini_.Use(secure.Secure(secure.Options{
		SSLRedirect: sslRedirect,
	}))

	martini_.Post(POST_URL, binding.Form(ProspectForm{}), errorHandler, createHandler)
	martini_.NotFound(notFoundHandler)
	martini_.Run()
}
