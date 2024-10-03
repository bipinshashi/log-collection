package internal

import (
	"bytes"
	"net/url"
	"reflect"
	"testing"
	"time"

	"github.com/bipinshashi/log-collection/internal/types"
)

func Test_parseLogEntry(t *testing.T) {
	type args struct {
		line    string
		logType types.LogEntryType
		server  string
	}
	tests := []struct {
		name string
		args args
		want types.LogEntry
	}{
		{
			name: "types.System log entry",
			args: args{
				line:    "Oct 1 13:08:11 This is a log entry",
				logType: types.System,
				server:  "api",
			},
			want: types.LogEntry{
				Timestamp: time.Date(0, time.October, 1, 13, 8, 11, 0, time.UTC),
				Server:    "api",
				Message:   "Oct 1 13:08:11 This is a log entry",
				Type:      types.System,
			},
		},
		{
			name: "Unparseable types.System log entry",
			args: args{
				line:    "WrongTimestamp This is a log entry",
				logType: types.System,
				server:  "api",
			},
			want: types.LogEntry{
				Timestamp: time.Time{},
				Server:    "",
				Message:   "",
				Type:      "",
			},
		},
		{
			name: "types.Wifi log entry",
			args: args{
				line:    "Mon Sep 30 10:27:25.955 This is a log entry",
				logType: types.Wifi,
				server:  "api",
			},
			want: types.LogEntry{
				Timestamp: time.Date(0, time.September, 30, 10, 27, 25, 955000000, time.UTC),
				Server:    "api",
				Message:   "Mon Sep 30 10:27:25.955 This is a log entry",
				Type:      types.Wifi,
			},
		},
		{
			name: "Unparseable types.Wifi log entry",
			args: args{
				line:    "Mon blurp Sep 30 10:27:25.955 This is a log entry",
				logType: types.Wifi,
				server:  "api",
			},
			want: types.LogEntry{
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
		want    []types.LogEntry
		wantErr bool
	}{
		{
			name: "Read last 1 line (n < total lines)",
			args: args{
				fileBytes: testLogReaderBytes,
				params:    RequestParams{fileName: "system.log", lines: 1},
				server:    "api",
			},
			want: []types.LogEntry{
				{
					Timestamp: time.Date(0, time.October, 1, 13, 8, 11, 0, time.UTC),
					Server:    "api",
					Message:   "Oct 1 13:08:11 This is a log entry",
					Type:      types.System,
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
			want: []types.LogEntry{
				{
					Timestamp: time.Date(0, time.October, 1, 13, 8, 11, 0, time.UTC),
					Server:    "api",
					Message:   "Oct 1 13:08:11 This is a log entry",
					Type:      types.System,
				},
				{
					Timestamp: time.Date(0, time.October, 1, 13, 8, 10, 0, time.UTC),
					Server:    "api",
					Message:   "Oct 1 13:08:10 This is another log entry",
					Type:      types.System,
				},
				{
					Timestamp: time.Date(0, time.October, 1, 13, 8, 9, 0, time.UTC),
					Server:    "api",
					Message:   "Oct 1 13:08:09 This is a log entry",
					Type:      types.System,
				},
				{
					Timestamp: time.Date(0, time.October, 1, 13, 8, 8, 0, time.UTC),
					Server:    "api",
					Message:   "Oct 1 13:08:08 This is another log entry",
					Type:      types.System,
				},
				{
					Timestamp: time.Date(0, time.October, 1, 13, 8, 7, 0, time.UTC),
					Server:    "api",
					Message:   "Oct 1 13:08:07 This is a log entry",
					Type:      types.System,
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
			want: []types.LogEntry{
				{
					Timestamp: time.Date(0, time.October, 1, 13, 8, 11, 0, time.UTC),
					Server:    "api",
					Message:   "Oct 1 13:08:11 This is a log entry",
					Type:      types.System,
				},
				{
					Timestamp: time.Date(0, time.October, 1, 13, 8, 10, 0, time.UTC),
					Server:    "api",
					Message:   "Oct 1 13:08:10 This is another log entry",
					Type:      types.System,
				},
				{
					Timestamp: time.Date(0, time.October, 1, 13, 8, 9, 0, time.UTC),
					Server:    "api",
					Message:   "Oct 1 13:08:09 This is a log entry",
					Type:      types.System,
				},
				{
					Timestamp: time.Date(0, time.October, 1, 13, 8, 8, 0, time.UTC),
					Server:    "api",
					Message:   "Oct 1 13:08:08 This is another log entry",
					Type:      types.System,
				},
				{
					Timestamp: time.Date(0, time.October, 1, 13, 8, 7, 0, time.UTC),
					Server:    "api",
					Message:   "Oct 1 13:08:07 This is a log entry",
					Type:      types.System,
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
			want: []types.LogEntry{
				{
					Timestamp: time.Date(0, time.October, 1, 13, 8, 10, 0, time.UTC),
					Server:    "api",
					Message:   "Oct 1 13:08:10 This is another log entry",
					Type:      types.System,
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

func Test_validateQueryParams(t *testing.T) {
	type args struct {
		url url.Values
	}
	tests := []struct {
		name    string
		args    args
		want    RequestParams
		wantErr bool
	}{
		{
			name: "Valid query params",
			args: args{
				url: url.Values{
					"file":   []string{"system.log"},
					"n":      []string{"5"},
					"filter": []string{"error"},
				},
			},
			want: RequestParams{
				fileName: "system.log",
				lines:    5,
				filter:   "error",
			},
			wantErr: false,
		},
		{
			name: "n is not a number",
			args: args{
				url: url.Values{
					"file":   []string{"system.log"},
					"n":      []string{"five"},
					"filter": []string{"error"},
				},
			},
			want:    RequestParams{},
			wantErr: true,
		},
		{
			name: "n is not provided, should default to 10",
			args: args{
				url: url.Values{
					"file":   []string{"system.log"},
					"filter": []string{"error"},
				},
			},
			want: RequestParams{
				fileName: "system.log",
				lines:    10,
				filter:   "error",
			},
			wantErr: false,
		},
		{
			name: "n is greater than 1000",
			args: args{
				url: url.Values{
					"file":   []string{"system.log"},
					"n":      []string{"1001"},
					"filter": []string{"error"},
				},
			},
			want:    RequestParams{},
			wantErr: true,
		},
		{
			name: "n is less than 1",
			args: args{
				url: url.Values{
					"file":   []string{"system.log"},
					"n":      []string{"0"},
					"filter": []string{"error"},
				},
			},
			want:    RequestParams{},
			wantErr: true,
		},
		{
			name: "n is negative",
			args: args{
				url: url.Values{
					"file":   []string{"system.log"},
					"n":      []string{"-5"},
					"filter": []string{"error"},
				},
			},
			want:    RequestParams{},
			wantErr: true,
		},
		{
			name: "filter is provided and has leading/trailing spaces",
			args: args{
				url: url.Values{
					"file":   []string{"system.log"},
					"n":      []string{"5"},
					"filter": []string{" Error "},
				},
			},
			want: RequestParams{
				fileName: "system.log",
				lines:    5,
				filter:   "error",
			},
			wantErr: false,
		},
		{
			name: "filter is not provided",
			args: args{
				url: url.Values{
					"file": []string{"system.log"},
					"n":    []string{"5"},
				},
			},
			want: RequestParams{
				fileName: "system.log",
				lines:    5,
				filter:   "",
			},
			wantErr: false,
		},
		{
			name: "file is not provided, should default to system.log",
			args: args{
				url: url.Values{
					"n":      []string{"5"},
					"filter": []string{"error"},
				},
			},
			want: RequestParams{
				fileName: "system.log",
				lines:    5,
				filter:   "error",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := validateQueryParams(tt.args.url)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateQueryParams() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("validateQueryParams() = %v, want %v", got, tt.want)
			}
		})
	}
}
