package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	var (
		port    = flag.Int("port", 8080, "Port to listen on")
		content = flag.String("content", "/posts", "Path to markdown posts")
		config  = flag.String("config", "", "Path to config file")
		help    = flag.Bool("help", false, "Show help message")
	)

	flag.Parse()

	if *help {
		fmt.Fprintf(os.Stdout, "GoBlogServ - Opinionated web server for markdown blogs\n\n")
		fmt.Fprintf(os.Stdout, "Version: %s\n", version)
		fmt.Fprintf(os.Stdout, "Commit:  %s\n", commit)
		fmt.Fprintf(os.Stdout, "Built:   %s\n\n", date)
		fmt.Fprintf(os.Stdout, "Usage:\n")
		flag.PrintDefaults()
		os.Exit(0)
	}

	// Placeholder implementation
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "GoBlogServ Placeholder\n\n")
		fmt.Fprintf(w, "Version: %s\n", version)
		fmt.Fprintf(w, "Content folder: %s\n", *content)
		if *config != "" {
			fmt.Fprintf(w, "Config: %s\n", *config)
		}
		fmt.Fprintf(w, "\nThis is a placeholder implementation.\n")
		fmt.Fprintf(w, "The full GoBlogServ server is not yet implemented.\n")
	})

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "OK")
	})

	addr := fmt.Sprintf(":%d", *port)
	log.Printf("GoBlogServ placeholder starting on %s", addr)
	log.Printf("Content folder: %s", *content)
	if *config != "" {
		log.Printf("Config: %s", *config)
	}
	log.Printf("Version: %s (%s)", version, commit)

	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatal(err)
	}
}
