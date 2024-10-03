package internal

import (
	"bytes"
	"reflect"
	"testing"
	"time"
)

func Test_parseLogEntry(t *testing.T) {
	type args struct {
		line    string
		logType LogEntryType
		server  string
	}
	tests := []struct {
		name string
		args args
		want LogEntry
	}{
		{
			name: "System log entry",
			args: args{
				line:    "Oct 1 13:08:11 This is a log entry",
				logType: System,
				server:  "api",
			},
			want: LogEntry{
				Timestamp: time.Date(0, time.October, 1, 13, 8, 11, 0, time.UTC),
				Server:    "api",
				Message:   "This is a log entry",
				Type:      System,
			},
		},
		{
			name: "Unparseable System log entry",
			args: args{
				line:    "WrongTimestamp This is a log entry",
				logType: System,
				server:  "api",
			},
			want: LogEntry{
				Timestamp: time.Time{},
				Server:    "",
				Message:   "",
				Type:      "",
			},
		},
		{
			name: "Wifi log entry",
			args: args{
				line:    "Mon Sep 30 10:27:25.955 This is a log entry",
				logType: Wifi,
				server:  "api",
			},
			want: LogEntry{
				Timestamp: time.Date(0, time.September, 30, 10, 27, 25, 955000000, time.UTC),
				Server:    "api",
				Message:   "This is a log entry",
				Type:      Wifi,
			},
		},
		{
			name: "Unparseable Wifi log entry",
			args: args{
				line:    "Mon blurp Sep 30 10:27:25.955 This is a log entry",
				logType: Wifi,
				server:  "api",
			},
			want: LogEntry{
				Timestamp: time.Time{},
				Server:    "",
				Message:   "",
				Type:      "",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := parseLogEntry(tt.args.line, tt.args.logType, tt.args.server); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseLogEntry() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_readLastNLines(t *testing.T) {
	type args struct {
		fileBytes []byte
		params    RequestParams
		server    string
	}
	testLogReaderBytes := []byte(`Oct 1 13:08:07 This is a log entry
	Oct 1 13:08:08 This is another log entry
	Oct 1 13:08:09 This is a log entry
	Oct 1 13:08:10 This is another log entry
	Oct 1 13:08:11 This is a log entry`)
	tests := []struct {
		name    string
		args    args
		want    []LogEntry
		wantErr bool
	}{
		{
			name: "Read last 1 line (n < total lines)",
			args: args{
				fileBytes: testLogReaderBytes,
				params:    RequestParams{fileName: "system.log", lines: 1},
				server:    "api",
			},
			want: []LogEntry{
				{
					Timestamp: time.Date(0, time.October, 1, 13, 8, 11, 0, time.UTC),
					Server:    "api",
					Message:   "Oct 1 13:08:11 This is a log entry",
					Type:      System,
				},
			},
			wantErr: false,
		},
		{
			name: "Read last 5 lines (n = total lines)",
			args: args{
				fileBytes: testLogReaderBytes,
				params:    RequestParams{fileName: "system.log", lines: 5},
				server:    "api",
			},
			want: []LogEntry{
				{
					Timestamp: time.Date(0, time.October, 1, 13, 8, 11, 0, time.UTC),
					Server:    "api",
					Message:   "Oct 1 13:08:11 This is a log entry",
					Type:      System,
				},
				{
					Timestamp: time.Date(0, time.October, 1, 13, 8, 10, 0, time.UTC),
					Server:    "api",
					Message:   "Oct 1 13:08:10 This is another log entry",
					Type:      System,
				},
				{
					Timestamp: time.Date(0, time.October, 1, 13, 8, 9, 0, time.UTC),
					Server:    "api",
					Message:   "Oct 1 13:08:09 This is a log entry",
					Type:      System,
				},
				{
					Timestamp: time.Date(0, time.October, 1, 13, 8, 8, 0, time.UTC),
					Server:    "api",
					Message:   "Oct 1 13:08:08 This is another log entry",
					Type:      System,
				},
				{
					Timestamp: time.Date(0, time.October, 1, 13, 8, 7, 0, time.UTC),
					Server:    "api",
					Message:   "Oct 1 13:08:07 This is a log entry",
					Type:      System,
				},
			},
			wantErr: false,
		},
		{
			name: "Read last 15 lines (n > total lines)",
			args: args{
				fileBytes: testLogReaderBytes,
				params:    RequestParams{fileName: "system.log", lines: 15},
				server:    "api",
			},
			want: []LogEntry{
				{
					Timestamp: time.Date(0, time.October, 1, 13, 8, 11, 0, time.UTC),
					Server:    "api",
					Message:   "Oct 1 13:08:11 This is a log entry",
					Type:      System,
				},
				{
					Timestamp: time.Date(0, time.October, 1, 13, 8, 10, 0, time.UTC),
					Server:    "api",
					Message:   "Oct 1 13:08:10 This is another log entry",
					Type:      System,
				},
				{
					Timestamp: time.Date(0, time.October, 1, 13, 8, 9, 0, time.UTC),
					Server:    "api",
					Message:   "Oct 1 13:08:09 This is a log entry",
					Type:      System,
				},
				{
					Timestamp: time.Date(0, time.October, 1, 13, 8, 8, 0, time.UTC),
					Server:    "api",
					Message:   "Oct 1 13:08:08 This is another log entry",
					Type:      System,
				},
				{
					Timestamp: time.Date(0, time.October, 1, 13, 8, 7, 0, time.UTC),
					Server:    "api",
					Message:   "Oct 1 13:08:07 This is a log entry",
					Type:      System,
				},
			},
			wantErr: false,
		},
		{
			name: "Read last 1 line with filter",
			args: args{
				fileBytes: testLogReaderBytes,
				params:    RequestParams{fileName: "system.log", lines: 1, filter: "another"},
				server:    "api",
			},
			want: []LogEntry{
				{
					Timestamp: time.Date(0, time.October, 1, 13, 8, 10, 0, time.UTC),
					Server:    "api",
					Message:   "Oct 1 13:08:10 This is another log entry",
					Type:      System,
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			file := bytes.NewReader(tt.args.fileBytes)
			got, err := readLastNLines(file, tt.args.params, tt.args.server)
			if (err != nil) != tt.wantErr {
				t.Errorf("readLastNLines() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("readLastNLines() = %v, want %v", got, tt.want)
			}
		})
	}
}
