package internal

import (
	"bufio"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/bipinshashi/log-collection/internal/config"
	"github.com/bipinshashi/log-collection/internal/utils"
)

type AppHandler struct {
	Client *http.Client
}

type RequestParams struct {
	fileName string
	lines    int
	filter   string
}

type LogEntryType string

const (
	System LogEntryType = "system"
	Wifi   LogEntryType = "wifi"
)

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

func (a *AppHandler) GetLogs(w http.ResponseWriter, r *http.Request) {
	nStr := r.URL.Query().Get("n")
	n := 100
	if nStr != "" {
		var err error
		n, err = strconv.Atoi(nStr)
		if err != nil {
			returnBadRequest("Invalid number of lines", w)
			return
		}
		if n <= 0 || n > 1000 {
			returnBadRequest("Number of lines must be between 1 and 1000", w)
			return
		}
	}

	// sanitize filter query
	filter := r.URL.Query().Get("filter")
	filter = strings.TrimSpace(filter)
	filter = strings.ToLower(filter)

	filename := r.URL.Query().Get("file")
	if filename == "" {
		filename = defaultLogFileName
	}
	params := RequestParams{
		fileName: filename,
		filter:   filter,
		lines:    n,
	}

	filePath, err := utils.ValidateFilePath(logDir, params.fileName)
	if err != nil {
		if os.IsNotExist(err) {
			returnBadRequest("File does not exist", w)
		} else if os.IsPermission(err) {
			returnBadRequest("Permission denied", w)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	file, err := os.Open(filePath)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer file.Close()

	server := config.GetConfig().ServerName
	logs, err := readLastNLines(file, params, server)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	logs = a.appendPeerLogs(logs, params)

	sort.Slice(logs, func(i, j int) bool {
		return logs[i].Timestamp.After(logs[j].Timestamp)
	})

	if len(logs) > params.lines {
		logs = logs[0:params.lines]
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(logs)
	w.WriteHeader(http.StatusOK)
}

func (a *AppHandler) appendPeerLogs(logs []LogEntry, params RequestParams) []LogEntry {
	config := config.GetConfig()
	if config.Peers == "" {
		return logs
	}

	// use a go channel to concurrently call peers
	jobs := make(chan string, len(strings.Split(config.Peers, ",")))
	peerLogResponses := make(chan []LogEntry, len(strings.Split(config.Peers, ",")))
	for _, peer := range strings.Split(config.Peers, ",") {
		url := getUrlForPeer(peer, params)
		jobs <- url
	}
	close(jobs)

	// worker pool to concurrently fetch logs from peers
	workercount := config.WorkerCount
	for w := 1; w <= workercount; w++ {
		go a.worker(jobs, peerLogResponses)
	}

	// wait for all peers to respond
	for i := 0; i < len(strings.Split(config.Peers, ",")); i++ {
		peerLogs := <-peerLogResponses
		logs = append(logs, peerLogs...)
	}
	return logs
}

func getUrlForPeer(peer string, params RequestParams) string {
	lines := params.lines
	linesStr := strconv.Itoa(lines)
	url := "http://" + peer + "/api/v1/logs?n=" + linesStr + "&file=" + params.fileName + "&filter=" + params.filter
	return url
}

func (a *AppHandler) worker(jobs <-chan string, peerLogResponses chan<- []LogEntry) {
	for url := range jobs {
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			log.Println(err)
			return
		}
		resp, err := a.Client.Do(req)
		if err != nil {
			log.Println(err)
			return
		}
		defer resp.Body.Close()
		var logs []LogEntry
		err = json.NewDecoder(resp.Body).Decode(&logs)
		if err != nil {
			log.Println(err)
			return
		}
		peerLogResponses <- logs
	}
}

func returnBadRequest(errorMsg string, w http.ResponseWriter) {
	http.Error(w, errorMsg, http.StatusBadRequest)
	w.WriteHeader(http.StatusBadRequest)
}

// readLastNLines reads the last n lines of a log file.
func readLastNLines(file io.Reader, params RequestParams, server string) ([]LogEntry, error) {
	// Read the file line by line
	scanner := bufio.NewScanner(file)
	var logs []LogEntry
	logType := getLogEntryType(params.fileName)

	for i := 0; scanner.Scan(); i++ {
		entry := parseLogEntry(scanner.Text(), logType, server)
		if params.filter != "" && strings.Contains(strings.ToLower(entry.Message), params.filter) {
			logs = append(logs, entry)
		} else if params.filter == "" {
			logs = append(logs, entry)
		}
	}
	sort.Slice(logs, func(i, j int) bool {
		return logs[i].Timestamp.After(logs[j].Timestamp)
	})
	if len(logs) < params.lines {
		return logs, nil
	}
	return logs[0:params.lines], nil
}

func getLogEntryType(filename string) LogEntryType {
	var logType LogEntryType
	if filename == "wifi.log" {
		logType = Wifi
	} else {
		logType = System
	}
	return logType
}

func parseLogEntry(line string, logType LogEntryType, server string) LogEntry {
	parts := strings.Fields(line)
	logEntryTypeConfig := LogEntryTypeTimePart[logType]
	if len(parts) < logEntryTypeConfig.Part {
		// increment a metric for failed log parsing
		return LogEntry{}
	}
	timestamp, err := time.Parse(logEntryTypeConfig.Layout, strings.Join(parts[0:logEntryTypeConfig.Part], " "))
	if err != nil {
		// increment a metric for failed log parsing
		return LogEntry{}
	}
	message := strings.Join(parts, " ")
	return LogEntry{
		Timestamp: timestamp,
		Server:    server,
		Message:   message,
		Type:      logType,
	}
}
