package internal

import (
	"fmt"
	"net/http"
	"strconv"
)

type RequestParams struct {
	fileName string
	lines    int
	filter   string
}

func GetLogs(w http.ResponseWriter, r *http.Request) {
	lines := r.URL.Query().Get("lines")
	if lines == "" {
		lines = "100"
	}
	params := RequestParams{
		fileName: r.URL.Query().Get("file"),
		filter:   r.URL.Query().Get("query"),
	}
	var err error
	params.lines, err = strconv.Atoi(lines)
	if err != nil {
		http.Error(w, "Invalid lines parameter", http.StatusBadRequest)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	fmt.Printf("File Name: %s\n", params.fileName)
	fmt.Printf("Lines: %d\n", params.lines)
	fmt.Printf("Query: %s\n", params.filter)
	w.Write([]byte("getting logs from file " + params.fileName))
	// Write the logs to the response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
}
