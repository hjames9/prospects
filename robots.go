package common

import (
	"fmt"
)

const (
	DISALLOW = "Disallow"
)

type RobotsRecord struct {
	UserAgents []string
	Disallows  []string
}

type RobotsTxt struct {
	RobotsRecords []RobotsRecord
}

func (robotsTxt *RobotsTxt) AddRecord(record RobotsRecord) {
	robotsTxt.RobotsRecords = append(robotsTxt.RobotsRecords, record)
}

func (robotsTxt RobotsTxt) String() string {
	var buffer string

	for _, robotsRecord := range robotsTxt.RobotsRecords {
		for _, userAgent := range robotsRecord.UserAgents {
			buffer += fmt.Sprintf("%s: %s\n", USER_AGENT_HEADER, userAgent)
		}

		for _, disallow := range robotsRecord.Disallows {
			buffer += fmt.Sprintf("%s: %s\n", DISALLOW, disallow)
		}

		buffer += "\n"
	}

	return buffer
}
