package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"sync"
	"time"

	pb "streakai/grpc"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type User struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type UserLogoutReq struct {
	Username string `json:"username"`
}

type VotingSession struct {
	Name     string   `json:"name"`
	Id       string   `json:"id"`
	YesCount []string `json:"yesCount"`
	NoCount  []string `json:"noCount"`
}

type SingleVote struct {
	Id   string `json:"id"`
	Vote bool   `json:"vote"`
}
type AllSessions []*VotingSession

var (
	sessions     AllSessions
	sessionMutex sync.Mutex
	grpcClient   pb.StreakAiServiceClient
	upgrader     = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			// Allowing all origins for now
			return true
		}}
	wsMutex sync.Mutex
	clients = make(map[*websocket.Conn]bool)
)

func initGRPCConnection() error {
	var err error
	conn, err := grpc.Dial("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return fmt.Errorf("failed to connect to gRPC server at localhost:50051: %v", err)
	}
	grpcClient = pb.NewStreakAiServiceClient(conn)
	return nil
}

func main() {

	sessions = make(AllSessions, 0)
	fmt.Println(sessions)

	err := initGRPCConnection()
	if err != nil {
		log.Fatalf("Error initializing gRPC connection: %v", err)
	}

	router := mux.NewRouter()
	router.HandleFunc("/login", handleLogin).Methods("POST")
	router.HandleFunc("/register", handleRegister).Methods("POST")
	router.HandleFunc("/logout", handleLogout).Methods("POST")
	router.HandleFunc("/ws", handleWebSocket)
	router.HandleFunc("/sessions", handleSessions)

	fmt.Println("Server is running at http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", router))
}

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
	case http.MethodPatch: //adds a new voting session
		castVote(w, r)
	case http.MethodGet: //fetches all voting session
		// getAllSessionHandler(w)

	default:
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
	}
}

func castVote(w http.ResponseWriter, r *http.Request) {

	log.Println("Handling cast vote request")
	var singleVote SingleVote
	defer r.Body.Close()
	if err := json.NewDecoder(r.Body).Decode(&singleVote); err != nil {
		log.Printf("Error decoding JSON: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	log.Printf("Decoded vote: %+v", singleVote)

	tokenString := r.Header.Get("Authorization")
	if tokenString == "" {
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprint(w, "Missing authorization header")
		return
	}
	tokenString = tokenString[len("Bearer "):]

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	resp, err := grpcClient.CheckAuthorized(ctx, &pb.CheckAuthorizedReq{AuthCode: tokenString})
	if err != nil {
		http.Error(w, "Not Authorized", http.StatusUnauthorized)
		log.Printf("gRPC authorization  failed: %v", err)
		return
	}

	for i, session := range sessions {
		log.Printf("%d: %+v", i, session)
		if session.Id == singleVote.Id {
			if singleVote.Vote {
				session.YesCount = append(session.YesCount, resp.Username)
				broadcastSessionStatus(session)
			}
		}
	}

}

func createSessionHandler(w http.ResponseWriter, r *http.Request) {

	log.Println("Handling create voting session request")
	var votingSession VotingSession
	defer r.Body.Close()
	if err := json.NewDecoder(r.Body).Decode(&votingSession); err != nil {
		log.Printf("Error decoding JSON: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	log.Printf("Decoded Job: %+v", votingSession)
	votingSession.Id = uuid.New().String()
	sessionMutex.Lock()
	sessions = append(sessions, &votingSession)
	sessionMutex.Unlock()
	broadcastSessionStatus(&votingSession)
	logAllSessions()
}

func logAllSessions() {
	sessionMutex.Lock()
	log.Println("Current job queue:")
	for i, session := range sessions {
		log.Printf("%d: %+v", i, session)
	}
	sessionMutex.Unlock()
}

func broadcastSessionStatus(session *VotingSession) {
	sessionMutex.Lock()
	data, err := json.Marshal(session)

	sessionMutex.Unlock()
	if err != nil {
		log.Printf("Error encoding session data: %v", err)
		sessionMutex.Unlock()
		return
	}
	for client := range clients {
		if err := client.WriteMessage(websocket.TextMessage, data); err != nil {
			log.Printf("Error writing message to client: %v", err)
			client.Close()
			delete(clients, client)
		}
	}
}

func handleLogin(w http.ResponseWriter, r *http.Request) {
	log.Println("Login handler hit")
	var user User

	body, err := ioutil.ReadAll(r.Body)

	if err != nil {
		http.Error(w, "Error reading request body", http.StatusInternalServerError)
		return
	}

	if err := json.Unmarshal(body, &user); err != nil {
		http.Error(w, "Error parsing request body", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	resp, err := grpcClient.Login(ctx, &pb.LoginRequest{Username: user.Username, Password: user.Password})
	if err != nil {
		http.Error(w, "Login failed", http.StatusUnauthorized)
		log.Printf("gRPC login call failed: %v", err)
		return
	}

	SendResponse(w, http.StatusOK, map[string]string{"token": resp.Token})
	log.Println("Login successful")
}

func handleRegister(w http.ResponseWriter, r *http.Request) {
	log.Println("Register handler hit")

	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Decode the request body
	var user User
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		log.Printf("Failed to decode request body: %v", err)
		return
	}
	log.Printf("User: %+v", user)

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	// Call the gRPC service for registration
	resp, err := grpcClient.Register(ctx, &pb.RegisterRequest{Username: user.Username, Password: user.Password})
	if err != nil {
		http.Error(w, "Registration failed", http.StatusInternalServerError)
		log.Printf("gRPC registration call failed: %v", err)
		return
	}

	// Send response
	SendResponse(w, http.StatusOK, map[string]string{"message": "Registration successful", "status": resp.Status})
	log.Println("Registration successful")
}

func handleLogout(w http.ResponseWriter, r *http.Request) {
	log.Println("LogOut handler hit")
	var userLogOutReq UserLogoutReq

	body, err := ioutil.ReadAll(r.Body)

	if err != nil {
		http.Error(w, "Error reading request body", http.StatusInternalServerError)
		return
	}

	if err := json.Unmarshal(body, &userLogOutReq); err != nil {
		http.Error(w, "Error parsing request body", http.StatusBadRequest)
		return
	}

	tokenString := r.Header.Get("Authorization")
	if tokenString == "" {
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprint(w, "Missing authorization header")
		return
	}
	tokenString = tokenString[len("Bearer "):]

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	resp, err := grpcClient.LogOut(ctx, &pb.LogOutRequest{Username: userLogOutReq.Username, AuthCode: tokenString})
	if err != nil {
		http.Error(w, "Login failed", http.StatusUnauthorized)
		log.Printf("gRPC login call failed: %v", err)
		return
	}
	SendResponse(w, http.StatusOK, map[string]string{"token": resp.Status})
	log.Println("Login successful")
}

func SendResponse(w http.ResponseWriter, statusCode int, payload interface{}) {
	response, err := json.Marshal(payload)
	if err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		log.Printf("Failed to encode response: %v", err)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(statusCode)
	w.Write(response)
}

func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	fmt.Println("web socket hit")
	// Set CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS,DELETE")
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
