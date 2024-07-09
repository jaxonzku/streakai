package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

func main() {

	sessions = make(AllSessions, 0)

	err := initGRPCConnection()
	if err != nil {
		log.Fatalf("Error initializing gRPC connection: %v", err)
	}
	initRedis()
	router := mux.NewRouter()
	router.HandleFunc("/login", handleLogin).Methods("POST")
	router.HandleFunc("/register", handleRegister).Methods("POST")
	router.HandleFunc("/logout", handleLogout).Methods("POST")
	router.HandleFunc("/ws", handleWebSocket)
	router.HandleFunc("/sessions", handleSessions)
	router.HandleFunc("/sessions/{id}", handleSessions)

	fmt.Println("Server is running at http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", router))
	fmt.Println("redis check", redisClient)
}
