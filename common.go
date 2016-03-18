package common

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"
	"time"
)

const (
	GET_METHOD        = "GET"
	USER_AGENT_HEADER = "User-Agent"
	USER_AGENT        = "Prospects"
)

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
