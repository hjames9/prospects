package common

import (
	"net/http"
)

const (
	BOT_ERROR = "BotError"
)

type RequestLocation int

const (
	Header RequestLocation = 1 << iota
	Body
)

type Position int

const (
	First Position = 1 << iota
	Last
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
