package main

import (
	"bitbucket.org/padium/prospects"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

type NumVerifyValidator struct {
	ApiKey string
}

func (validator NumVerifyValidator) Validate(prospect common.Prospect) (bool, bool, string) {
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

	if len(prospect.PhoneNumber) <= 0 {
		log.Printf("No phone number to validate id %d", prospect.Id)
		return isValid, wasProcessed, miscellaneous
	}

	url := fmt.Sprintf(URL, validator.ApiKey, prospect.PhoneNumber)
	body, responseCode, _, err = common.MakeHttpGetRequest(url)
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
