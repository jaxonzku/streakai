package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	fmt.Println("WebSocket hit")

	// Set CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS, DELETE")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	// Handle OPTIONS request
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Upgrade error: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	wsMutex.Lock()
	clients[conn] = true
	wsMutex.Unlock()

	log.Println("WebSocket connection established")
}

func logAllSessions() {
	sessionMutex.Lock()
	defer sessionMutex.Unlock()
	log.Println("Current job queue:")
	for i, session := range sessions {
		log.Printf("%d: %+v", i, session)
	}
}

func broadcastSessionStatus(session *VotingSession) {
	sessionMutex.Lock()
	data, err := json.Marshal(session)
	sessionMutex.Unlock()
	if err != nil {
		log.Printf("Error encoding session data: %v", err)
		return
	}

	for client := range clients {
		if err := client.WriteMessage(websocket.TextMessage, data); err != nil {
			log.Printf("Error writing message to client: %v", err)
			client.Close()
			wsMutex.Lock()
			delete(clients, client)
			wsMutex.Unlock()
		}
	}
}
