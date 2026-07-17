package main

import (
	"embed"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"
)

//go:embed index.html
var indexHTML string

//go:embed style.css
var staticFS embed.FS

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
	mux.Handle("/style.css", http.FileServer(http.FS(staticFS)))
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "ok")
	})
	return mux
}

// pageData is passed to the HTML template.
type pageData struct {
	Hosts   []Host
	Message string
	IsError bool
}

func (s *server) handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	s.render(w, pageData{Hosts: s.cfg.Hosts})
}

func (s *server) handleWake(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	idx, err := strconv.Atoi(r.FormValue("host"))
	if err != nil || idx < 0 || idx >= len(s.cfg.Hosts) {
		s.render(w, pageData{Hosts: s.cfg.Hosts, Message: "Unknown host.", IsError: true})
		return
	}

	h := s.cfg.Hosts[idx]
	if err := sendMagicPacket(h.hw, h.Broadcast, s.cfg.WolPort); err != nil {
		log.Printf("wake %q failed: %v", h.Name, err)
		s.render(w, pageData{
			Hosts:   s.cfg.Hosts,
			Message: fmt.Sprintf("Failed to wake %s: %v", h.Name, err),
			IsError: true,
		})
		return
	}

	log.Printf("sent magic packet to %q (%s) via %s:%d", h.Name, h.MAC, h.Broadcast, s.cfg.WolPort)
	s.render(w, pageData{
		Hosts:   s.cfg.Hosts,
		Message: fmt.Sprintf("Magic packet sent to %s.", h.Name),
	})
}

func (s *server) render(w http.ResponseWriter, data pageData) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := pageTmpl.Execute(w, data); err != nil {
		log.Printf("template error: %v", err)
	}
}
