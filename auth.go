package main

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	pb "streakai/grpc"
)

// handleLogin processes login requests
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

// handleRegister processes registration requests
func handleRegister(w http.ResponseWriter, r *http.Request) {
	log.Println("Register handler hit")

	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var user User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		log.Printf("Failed to decode request body: %v", err)
		return
	}
	log.Printf("User: %+v", user)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	resp, err := grpcClient.Register(ctx, &pb.RegisterRequest{Username: user.Username, Password: user.Password})
	if err != nil {
		http.Error(w, "Registration failed", http.StatusInternalServerError)
		log.Printf("gRPC registration call failed: %v", err)
		return
	}

	SendResponse(w, http.StatusOK, map[string]string{"message": "Registration successful", "status": resp.Status})
	log.Println("Registration successful")
}

// handleLogout processes logout requests
func handleLogout(w http.ResponseWriter, r *http.Request) {
	log.Println("Logout handler hit")
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
		http.Error(w, "Missing authorization header", http.StatusUnauthorized)
		return
	}
	tokenString = tokenString[len("Bearer "):]

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	resp, err := grpcClient.LogOut(ctx, &pb.LogOutRequest{Username: userLogOutReq.Username, AuthCode: tokenString})
	if err != nil {
		http.Error(w, "Logout failed", http.StatusUnauthorized)
		log.Printf("gRPC logout call failed: %v", err)
		return
	}

	SendResponse(w, http.StatusOK, map[string]string{"status": resp.Status})
	log.Println("Logout successful")
}

// isAuthorised checks if the user is authorized
func isAuthorised(w http.ResponseWriter, r *http.Request) (string, bool) {
	tokenString := r.Header.Get("Authorization")
	if tokenString == "" {
		http.Error(w, "Missing authorization header", http.StatusUnauthorized)
		return "", false
	}
	tokenString = tokenString[len("Bearer "):]

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	resp, err := grpcClient.CheckAuthorized(ctx, &pb.CheckAuthorizedReq{AuthCode: tokenString})
	if err != nil || !resp.Authorized {
		http.Error(w, "Not Authorized", http.StatusUnauthorized)
		log.Printf("gRPC authorization failed: %v", err)
		return "", false
	}
	return resp.Username, true
}
