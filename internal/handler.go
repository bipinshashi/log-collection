package internal

import (
	"bufio"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/a-h/templ"
	"github.com/bipinshashi/log-collection/internal/components"
	"github.com/bipinshashi/log-collection/internal/config"
	"github.com/bipinshashi/log-collection/internal/types"
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

const (
	logDir             = "/var/log/"
	defaultLogFileName = "system.log"
	defaultLines       = 10
)

var LogEntryTypeTimePart = map[types.LogEntryType]types.LogEntryTypeConfig{
	types.System: {Layout: "Jan 2 15:04:05", Part: 3},
	types.Wifi:   {Layout: "Mon Jan 2 15:04:05.000", Part: 4},
}

var globalLogs types.GlobalLogState

func (a *AppHandler) ShowDemo(w http.ResponseWriter, r *http.Request) {
	// Update state.
	r.ParseForm()
	logs, shouldReturn := a.getLogsHelper(r, w)
	if shouldReturn {
		return
	}
	globalLogs.Entries = logs
	component := components.Page(globalLogs)
	templ.Handler(component).ServeHTTP(w, r)
}

func (a *AppHandler) GetLogs(w http.ResponseWriter, r *http.Request) {
	// update global logs
	logs, shouldReturn := a.getLogsHelper(r, w)
	if shouldReturn {
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(logs)
	w.WriteHeader(http.StatusOK)
}

func (a *AppHandler) getLogsHelper(r *http.Request, w http.ResponseWriter) ([]types.LogEntry, bool) {
	params, err := validateQueryParams(r.URL.Query())
	if err != nil {
		returnBadRequest(err.Error(), w)
		return nil, true
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
		return nil, true
	}

	file, err := os.Open(filePath)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		w.WriteHeader(http.StatusInternalServerError)
		return nil, true
	}
	defer file.Close()

	server := config.GetConfig().ServerName
	logs, err := readLastNLines(file, params, server)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		w.WriteHeader(http.StatusInternalServerError)
		return nil, true
	}

	logs = a.appendPeerLogs(logs, params)

	sort.Slice(logs, func(i, j int) bool {
		return logs[i].Timestamp.After(logs[j].Timestamp)
	})

	if len(logs) > params.lines {
		logs = logs[0:params.lines]
	}

	return logs, false
}

func (a *AppHandler) appendPeerLogs(logs []types.LogEntry, params RequestParams) []types.LogEntry {
	config := config.GetConfig()
	if config.Peers == "" {
		return logs
	}

	// use a go channel to concurrently call peers
	jobs := make(chan string, len(strings.Split(config.Peers, ",")))
	peerLogResponses := make(chan []types.LogEntry, len(strings.Split(config.Peers, ",")))
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

func validateQueryParams(values url.Values) (RequestParams, error) {
	nStr := values.Get("n")
	n := defaultLines
	if nStr != "" {
		var err error
		n, err = strconv.Atoi(nStr)
		if err != nil {
			return RequestParams{}, errors.New("invalid number of lines")
		}
		if n <= 0 || n > 1000 {
			return RequestParams{}, errors.New("number of lines should be between 1 and 1000")
		}
	}

	// sanitize filter query
	filter := values.Get("filter")
	filter = strings.TrimSpace(filter)
	filter = strings.ToLower(filter)

	filename := values.Get("file")
	if filename == "" {
		filename = defaultLogFileName
	}
	params := RequestParams{
		fileName: filename,
		filter:   filter,
		lines:    n,
	}

	return params, nil
}

func getUrlForPeer(peer string, params RequestParams) string {
	lines := params.lines
	linesStr := strconv.Itoa(lines)
	url := "http://" + peer + "/api/v1/logs?n=" + linesStr + "&file=" + params.fileName + "&filter=" + params.filter
	return url
}

func (a *AppHandler) worker(jobs <-chan string, peerLogResponses chan<- []types.LogEntry) {
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
		var logs []types.LogEntry
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
func readLastNLines(file io.Reader, params RequestParams, server string) ([]types.LogEntry, error) {
	// Read the file line by line
	scanner := bufio.NewScanner(file)
	var logs []types.LogEntry
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

func getLogEntryType(filename string) types.LogEntryType {
	var logType types.LogEntryType
	switch {
	case strings.Contains(filename, "wifi"):
		logType = types.Wifi
	case strings.Contains(filename, "system"):
		logType = types.System
	default:
		logType = types.System
	}
	return logType
}

func parseLogEntry(line string, logType types.LogEntryType, server string) types.LogEntry {
	parts := strings.Fields(line)
	logEntryTypeConfig := types.LogEntryTypeTimePart[logType]
	if len(parts) < logEntryTypeConfig.Part {
		return types.LogEntry{}
	}
	timestamp, err := time.Parse(logEntryTypeConfig.Layout, strings.Join(parts[0:logEntryTypeConfig.Part], " "))
	if err != nil {
		return types.LogEntry{}
	}
	message := strings.Join(parts, " ")
	return types.LogEntry{
		Timestamp: timestamp,
		Server:    server,
		Message:   message,
		Type:      logType,
	}
}
