package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
)

func main() {
	cfgPath := os.Getenv("CONFIG_PATH")
	if cfgPath == "" {
		cfgPath = defaultCfgPath
	}

	cfg, err := loadConfig(cfgPath)
	if err != nil {
		log.Fatalf("failed to load config %q: %v", cfgPath, err)
	}

	// PORT env overrides the config value.
	port := cfg.Port
	if p := os.Getenv("PORT"); p != "" {
		n, err := strconv.Atoi(p)
		if err != nil {
			log.Fatalf("invalid PORT %q: %v", p, err)
		}
		port = n
	}

	addr := fmt.Sprintf(":%d", port)
	log.Printf("pc-waker listening on %s with %d host(s)", addr, len(cfg.Hosts))
	if err := http.ListenAndServe(addr, newServer(cfg).routes()); err != nil {
		log.Fatal(err)
	}
}
