package common

import (
	"bytes"
	"database/sql"
	"fmt"
	"gopkg.in/gomail.v2"
	"html/template"
	"log"
	"net/url"
	"strconv"
	"strings"
)

const (
	TO_HEADER         = "To"
	SUBJECT_HEADER    = "Subject"
	HTML_CONTENT_TYPE = "text/html"
	MAILER_USER_AGENT = "James Mailer"
)

type ProcessCallback interface {
	Processed(map[string]string, string, bool) bool
}

type DatabaseTemplateMailer struct {
	SmtpServer         string
	SmtpUser           string
	SmtpPassword       string
	SmtpSubject        string
	SmtpSubjectValues  []interface{}
	SmtpTemplateUrl    string
	DatabaseQuery      string
	DatabaseParameters []interface{}
	DestEmailColumn    string
	SourceEmail        string
	DatabaseConnection *sql.DB
	Callback           ProcessCallback
}

func (dtm *DatabaseTemplateMailer) getHtmlTemplate() (*template.Template, error) {
	//Get URL
	smtpTemplateUrl, err := url.Parse(dtm.SmtpTemplateUrl)
	if nil != err {
		return nil, err
	} else if !strings.HasPrefix(smtpTemplateUrl.Scheme, "http") {
		return nil, fmt.Errorf("SMTP template URL is not http based: %s", smtpTemplateUrl.Scheme)
	}

	//Get HTML template
	smtpTemplate, responseCode, _, err := MakeHttpGetRequest(smtpTemplateUrl.String())
	if nil != err {
		return nil, err
	}

	if responseCode < 200 || responseCode > 299 || len(smtpTemplate) <= 0 {
		return nil, fmt.Errorf("Could not retrieve SMTP template.  Status code %d", responseCode)
	}

	//HTML templating
	tmpl, err := template.New("foo").Parse(string(smtpTemplate))
	if nil != err {
		return nil, err
	}

	return tmpl, err
}

func (dtm *DatabaseTemplateMailer) connectToSmtpServer() (gomail.SendCloser, error) {
	//Get smtp server details
	var (
		smtpServer string
		smtpPort   int
		err        error
	)

	smtpPair := strings.Split(dtm.SmtpServer, ":")
	smtpPort = 25
	if len(smtpPair) == 2 {
		smtpPort, err = strconv.Atoi(smtpPair[1])
		smtpServer = smtpPair[0]
		if nil != err {
			log.Printf("Invalid port number specified: %s.  Setting to default port 25.", smtpPair[1])
			log.Print(err)
			smtpPort = 25
		}
	}

	//SMTP client
	smtpClient := gomail.NewDialer(smtpServer, smtpPort, dtm.SmtpUser, dtm.SmtpPassword)
	sender, err := smtpClient.Dial()
	if nil != err {
		return nil, err
	}

	return sender, err
}

func (dtm *DatabaseTemplateMailer) getTemplateDataFromDatabase() ([]map[string]string, error) {
	rows, err := dtm.DatabaseConnection.Query(dtm.DatabaseQuery, dtm.DatabaseParameters...)
	if nil != err {
		return nil, err
	}
	defer rows.Close()

	cols, err := rows.Columns()
	data := make([]map[string]string, 0)

	databaseFieldsStr := make([]sql.NullString, len(cols))
	databaseFields := make([]interface{}, 0)

	for iter := 0; iter < len(cols); iter++ {
		databaseFields = append(databaseFields, &databaseFieldsStr[iter])
	}

	for rows.Next() {
		err = rows.Scan(databaseFields...)
		if nil != err {
			continue
		}

		databaseValues := make(map[string]string, len(cols))
		for iter, element := range databaseFields {
			nullString := *element.(*sql.NullString)
			if nullString.Valid {
				databaseValues[cols[iter]] = nullString.String
			} else {
				databaseValues[cols[iter]] = ""
			}
		}

		data = append(data, databaseValues)
	}

	if nil == err {
		err = rows.Err()
	}

	return data, err
}

func (dtm *DatabaseTemplateMailer) SendMail() error {
	//HTML template
	htmlTemplate, err := dtm.getHtmlTemplate()
	if nil != err {
		return err
	}

	//SMTP client
	sender, err := dtm.connectToSmtpServer()
	if nil != err {
		return err
	}
	defer sender.Close()

	//Template data
	templateDatas, err := dtm.getTemplateDataFromDatabase()
	if nil != err {
		return err
	}

	var tmplBuffer bytes.Buffer
	for _, templateData := range templateDatas {
		err = htmlTemplate.Execute(&tmplBuffer, templateData)
		if nil != err {
			return err
		}

		smtpSubjectValues := make([]string, 0)
		for _, smtpSubjectValue := range dtm.SmtpSubjectValues {
			if _, exists := templateData[smtpSubjectValue.(string)]; exists {
				smtpSubjectValues = append(smtpSubjectValues, templateData[smtpSubjectValue.(string)])
			} else {
				log.Printf("SMTP subject value %s does not exist in template data", smtpSubjectValue.(string))
			}
		}

		emailSubject := fmt.Sprintf(dtm.SmtpSubject, smtpSubjectValues)

		message := gomail.NewMessage()
		message.SetHeader(FROM_HEADER, dtm.SourceEmail)
		message.SetHeader(TO_HEADER, templateData[dtm.DestEmailColumn])
		message.SetHeader(SUBJECT_HEADER, emailSubject)
		message.SetHeader(USER_AGENT_HEADER, MAILER_USER_AGENT)
		message.SetBody(HTML_CONTENT_TYPE, tmplBuffer.String())

		err = sender.Send(dtm.SourceEmail, []string{templateData[dtm.DestEmailColumn]}, message)
		var success bool

		if nil != err {
			log.Print(err)
			success = false
		} else {
			success = true
		}

		if nil != dtm.Callback && !dtm.Callback.Processed(templateData, tmplBuffer.String(), success) {
			break
		}
	}

	return nil
}
