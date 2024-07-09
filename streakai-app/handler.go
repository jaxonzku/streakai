package main

import (
	"log"
	"net/http"
)

func handleSessions(w http.ResponseWriter, r *http.Request) {
	// CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS,DELETE")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	// Handle OPTIONS request
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	log.Printf("Session handler  hit: %s %s", r.Method, r.URL.Path)

	switch r.Method {
	case http.MethodPost: //adds a new voting session
		createSessionHandler(w, r)
	case http.MethodPatch: //cast a vote to a session
		castVote(w, r)
	case http.MethodGet: //fetches voting session
		getSessionHandler(w, r)

	default:
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
	}
}
