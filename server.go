package main

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"
)

const (
	readHeaderTimeout = 10 * time.Second
	rootStatusJSON    = `{"status":"ok","name":"seoscore-cli-server"}`
	scanCountJSON     = `{"count":2847}`
	healthJSON        = `{"status":"ok"}`
	jsonContentType   = "application/json"
)

type scanRequest struct {
	URL       string `json:"url"`
	Keyphrase string `json:"keyphrase"`
}

func newMux(client *http.Client, remote bool) *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/", handleRoot)
	mux.HandleFunc("/api/scan", scanHandler(client, remote))
	mux.HandleFunc("/api/scan-count", handleScanCount)
	mux.HandleFunc("/health", handleHealth)
	return mux
}

func newServer(addr string, client *http.Client, remote bool) *http.Server {
	return &http.Server{Addr: addr, Handler: newMux(client, remote), ReadHeaderTimeout: readHeaderTimeout}
}

func allowAnyOrigin(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
}

func writeBody(w http.ResponseWriter, body string) {
	_, _ = w.Write([]byte(body))
}

func handleRoot(w http.ResponseWriter, r *http.Request) {
	allowAnyOrigin(w)
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	if strings.Contains(r.Header.Get("Accept"), "text/html") {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		writeBody(w, localWebUIHTML)
		return
	}
	w.Header().Set("Content-Type", jsonContentType)
	writeBody(w, rootStatusJSON)
}

func scanHandler(client *http.Client, remote bool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		allowAnyOrigin(w)
		w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		switch r.Method {
		case http.MethodOptions:
			w.WriteHeader(http.StatusOK)
			return
		case http.MethodPost:
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var req scanRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.URL == "" {
			http.Error(w, "bad request: missing url", http.StatusBadRequest)
			return
		}
		res, err := scanTarget(client, req.URL, req.Keyphrase, remote)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", jsonContentType)
		_ = json.NewEncoder(w).Encode(res)
	}
}

func handleScanCount(w http.ResponseWriter, _ *http.Request) {
	allowAnyOrigin(w)
	w.Header().Set("Content-Type", jsonContentType)
	writeBody(w, scanCountJSON)
}

func handleHealth(w http.ResponseWriter, _ *http.Request) {
	allowAnyOrigin(w)
	writeBody(w, healthJSON)
}
