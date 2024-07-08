package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	pb "streakai/grpc"

	"github.com/gorilla/mux"
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

var (
	grpcClient pb.StreakAiServiceClient
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
	err := initGRPCConnection()
	if err != nil {
		log.Fatalf("Error initializing gRPC connection: %v", err)
	}

	router := mux.NewRouter()
	router.HandleFunc("/login", handleLogin).Methods("POST")
	router.HandleFunc("/register", handleRegister).Methods("POST")
	router.HandleFunc("/logout", handleLogout).Methods("POST")

	fmt.Println("Server is running at http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", router))
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
