package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	pb "streakai/grpc"
	"time"
)

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

func isAuthorised(w http.ResponseWriter, r *http.Request) (string, bool) {
	tokenString := r.Header.Get("Authorization")
	if tokenString == "" {
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprint(w, "Missing authorization header")
		return "", false
	}
	tokenString = tokenString[len("Bearer "):]

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	resp, err := grpcClient.CheckAuthorized(ctx, &pb.CheckAuthorizedReq{AuthCode: tokenString})
	if err != nil {
		http.Error(w, "Not Authorized", http.StatusUnauthorized)
		log.Printf("gRPC authorization  failed: %v", err)
		return "", false
	}
	return resp.Username, resp.Authorized
}
