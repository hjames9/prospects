package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

type NumVerifyValidator struct {
	ApiKey string
}

func (validator NumVerifyValidator) Validate(prospect Prospect) (bool, bool, string) {
	const (
		URL = "http://apilayer.net/api/validate?access_key=%s&number=%s"
	)

	var (
		body          []byte
		responseCode  int
		err           error
		miscellaneous string
	)

	isValid := false
	wasProcessed := false

	if !prospect.phoneNumber.Valid {
		log.Printf("No phone number to validate id %d", prospect.id)
		return isValid, wasProcessed, miscellaneous
	}

	url := fmt.Sprintf(URL, validator.ApiKey, prospect.phoneNumber.String)
	body, responseCode, err = MakeHttpGetRequest(url)
	if nil != err {
		return isValid, wasProcessed, miscellaneous
	} else {
		type Message struct {
			Valid bool
		}
		var message Message
		err = json.Unmarshal(body, &message)
		if nil == err {
			wasProcessed = responseCode == http.StatusOK
			isValid = message.Valid
			miscellaneous = string(body)
		} else {
			log.Print("Error processing json message: " + string(body))
			log.Print(err)
		}
	}

	return isValid, wasProcessed, miscellaneous
}
