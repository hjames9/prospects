package common

import (
	"database/sql"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"
	"time"
)

const (
	GET_METHOD        = "GET"
	POST_METHOD       = "POST"
	HEAD_METHOD       = "HEAD"
	USER_AGENT_HEADER = "User-Agent"
	USER_AGENT        = "Prospects"
	DB_DRIVER         = "postgres"
	FROM_HEADER       = "From"
	XFP_HEADER        = "X-Forwarded-Proto"
)

type Prospect struct {
	Id            int64
	LeadId        string `form:"leadid"`
	AppName       string `form:"appname" binding:"required"`
	Referrer      string
	PageReferrer  string `form:"pagereferrer"`
	FirstName     string `form:"firstname"`
	LastName      string `form:"lastname"`
	Email         string `form:"email"`
	LeadSource    string `form:"leadsource" binding:"required"`
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
	WasProcessed  bool
	IsValid       bool
}

type Response struct {
	Code    int
	Message string
	Id      int64 `json:",omitempty"`
}

func IsJSON(str string) bool {
	var js map[string]interface{}
	return json.Unmarshal([]byte(str), &js) == nil
}

func GetAge(timeVal time.Time) int64 {
	age := time.Now().Sub(timeVal).Seconds() / 31536000
	return int64(age)
}

func GetenvWithDefault(envKey string, defaultVal string) string {
	envVal := os.Getenv(envKey)

	if len(envVal) == 0 {
		envVal = defaultVal
	}

	return envVal
}

func MakeHttpGetRequest(url string) ([]byte, int, map[string][]string, error) {
	//Create HTTP client
	client := http.Client{}

	//Create request with headers
	request, err := http.NewRequest(GET_METHOD, url, nil)
	request.Header.Add(USER_AGENT_HEADER, USER_AGENT)
	if nil != err {
		return nil, 0, nil, err
	}

	//Execute request
	response, err := client.Do(request)
	if nil != err {
		return nil, 0, nil, err
	}

	//Get response body
	body, err := ioutil.ReadAll(response.Body)
	response.Body.Close()
	if nil != err {
		return nil, 0, nil, err
	}

	return body, response.StatusCode, response.Header, nil
}

func GetProspects(db *sql.DB, query string, args ...interface{}) ([]Prospect, error) {
	const (
		QUERY = "SELECT id, lead_id, lead_source, app_name, email, phone_number, miscellaneous, was_processed, is_valid "
	)

	rows, err := db.Query(QUERY+query, args...)
	if nil != err {
		return nil, err
	}
	defer rows.Close()

	var (
		id            int64
		leadId        string
		leadSource    string
		appName       string
		email         sql.NullString
		phoneNumber   sql.NullString
		miscellaneous sql.NullString
		wasProcessed  bool
		isValid       bool
	)

	var prospects []Prospect

	for rows.Next() {
		err := rows.Scan(&id, &leadId, &leadSource, &appName, &email, &phoneNumber, &miscellaneous, &wasProcessed, &isValid)
		if nil != err {
			continue
		}

		var prospect Prospect

		prospect.Id = id
		prospect.LeadId = leadId
		prospect.LeadSource = leadSource
		prospect.AppName = appName
		prospect.Email = email.String
		prospect.PhoneNumber = phoneNumber.String
		prospect.Miscellaneous = miscellaneous.String
		prospect.WasProcessed = wasProcessed
		prospect.IsValid = isValid

		prospects = append(prospects, prospect)
	}

	err = rows.Err()
	if nil != err {
		return prospects, err
	}

	return prospects, nil
}

func GetScheme(request *http.Request) string {
	prot := request.Header.Get(XFP_HEADER)
	if len(prot) > 0 {
		return prot
	}

	if nil == request.TLS {
		return "http"
	} else {
		return "https"
	}
}
