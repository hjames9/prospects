package main

import (
	"bitbucket.org/padium/prospects"
	"fmt"
	"log"
	"net/http"
)

type FullContactValidator struct {
	ApiKey string
}

func (validator FullContactValidator) Validate(prospect common.Prospect) (bool, bool, string) {
	const (
		URL = "https://api.fullcontact.com/v2/person.json?apiKey=%s&email=%s"
	)

	var (
		body          []byte
		responseCode  int
		err           error
		miscellaneous string
	)

	isValid := false
	wasProcessed := false

	if len(prospect.Email) <= 0 {
		log.Printf("No e-mail to validate id %d", prospect.Id)
		return isValid, wasProcessed, miscellaneous
	}

	url := fmt.Sprintf(URL, validator.ApiKey, prospect.Email)
	body, responseCode, _, err = common.MakeHttpGetRequest(url)
	if nil != err {
		log.Print("Error retrieving data from FullContact")
		log.Print(err)
		return isValid, wasProcessed, miscellaneous
	} else {
		wasProcessed = (responseCode == http.StatusOK) || (responseCode == http.StatusNotFound)
		isValid = (responseCode == http.StatusOK)
		miscellaneous = string(body)
	}

	return isValid, wasProcessed, miscellaneous
}
