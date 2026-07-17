package main

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"net/url"
	"strconv"
)

//go:embed web/index.html
var indexHTML string

//go:embed web/style.css
var styleCSS string

var pageTmpl = template.Must(template.New("page").Parse(indexHTML))

// server serves the web UI and handles wake requests.
type server struct {
	cfg *Config
}

func newServer(cfg *Config) *server {
	return &server{cfg: cfg}
}

// routes builds the HTTP handler for the server.
func (s *server) routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/", s.handleIndex)
	mux.HandleFunc("/wake", s.handleWake)
	mux.HandleFunc("/style.css", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/css; charset=utf-8")
		fmt.Fprint(w, styleCSS)
	})
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "ok")
	})
	return mux
}

// pageData is passed to the HTML template.
type pageData struct {
	Hosts []Host
}

func (s *server) handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	s.render(w, pageData{Hosts: s.cfg.Hosts})
}

// wakeResult is the JSON returned to the fetch() call in the page.
type wakeResult struct {
	Message string `json:"message"`
	Error   bool   `json:"error"`
}

func (s *server) handleWake(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if !sameOrigin(r) {
		http.Error(w, "cross-origin request rejected", http.StatusForbidden)
		return
	}

	idx, err := strconv.Atoi(r.FormValue("host"))
	if err != nil || idx < 0 || idx >= len(s.cfg.Hosts) {
		writeJSON(w, http.StatusBadRequest, wakeResult{Message: "Unknown host.", Error: true})
		return
	}

	h := s.cfg.Hosts[idx]
	if err := sendMagicPacket(h.hw, h.Broadcast, s.cfg.WolPort); err != nil {
		log.Printf("wake %q failed: %v", h.Name, err)
		writeJSON(w, http.StatusInternalServerError, wakeResult{
			Message: fmt.Sprintf("Failed to wake %s: %v", h.Name, err),
			Error:   true,
		})
		return
	}

	log.Printf("sent magic packet to %q (%s) via %s:%d", h.Name, h.MAC, h.Broadcast, s.cfg.WolPort)
	writeJSON(w, http.StatusOK, wakeResult{Message: fmt.Sprintf("Magic packet sent to %s.", h.Name)})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("json encode error: %v", err)
	}
}

func (s *server) render(w http.ResponseWriter, data pageData) {
	var buf bytes.Buffer
	if err := pageTmpl.Execute(&buf, data); err != nil {
		log.Printf("template error: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if _, err := buf.WriteTo(w); err != nil {
		log.Printf("write error: %v", err)
	}
}

// sameOrigin guards state-changing requests against CSRF. A cross-origin page
// that submits to /wake will carry an Origin (and usually Referer) header
// identifying that page; we reject when it names a host other than our own.
// Requests with neither header (e.g. some non-browser clients) are allowed
// through, since the CSRF vector requires a browser that sets these headers.
func sameOrigin(r *http.Request) bool {
	for _, h := range []string{r.Header.Get("Origin"), r.Header.Get("Referer")} {
		if h == "" {
			continue
		}
		u, err := url.Parse(h)
		if err != nil {
			return false
		}
		return u.Host == r.Host
	}
	return true
}
