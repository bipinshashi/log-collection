package types

import "time"

type LogEntryType string

const (
	System LogEntryType = "system"
	Wifi   LogEntryType = "wifi"
)

type GlobalLogState struct {
	Entries []LogEntry
}

type LogEntry struct {
	Timestamp time.Time    `json:"timestamp"`
	Server    string       `json:"server"`
	Message   string       `json:"message"`
	Type      LogEntryType `json:"type"`
}

const (
	logDir             = "/var/log/"
	defaultLogFileName = "system.log"
)

type LogEntryTypeConfig struct {
	Layout string
	Part   int
}

var LogEntryTypeTimePart = map[LogEntryType]LogEntryTypeConfig{
	System: {Layout: "Jan 2 15:04:05", Part: 3},
	Wifi:   {Layout: "Mon Jan 2 15:04:05.000", Part: 4},
}
