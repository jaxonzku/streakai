package main

import (
	"context"
	"fmt"
	"log"
	"time"

	pb "streakauth/grpc"

	"github.com/golang-jwt/jwt"
)

// Login handles user login requests
func (s *server) Login(ctx context.Context, in *pb.LoginRequest) (*pb.LoginResponse, error) {
	log.Printf("Received Login request: %v", in)

	if !isUserRegistered(in.Username) {
		log.Printf("Username not found: %s", in.Username)
		return &pb.LoginResponse{Token: ""}, fmt.Errorf("username not found")
	}

	if registeredUsers[in.Username] != in.Password {
		log.Printf("Invalid password for username: %s", in.Username)
		return &pb.LoginResponse{Token: ""}, fmt.Errorf("invalid password")
	}

	tokenString, err := CreateToken(in.Username)
	if err != nil {
		log.Printf("Error generating token: %v", err)
		return &pb.LoginResponse{Token: ""}, fmt.Errorf("error generating token")
	}

	loggedinUsers = append(loggedinUsers, in.Username)
	log.Printf("Logged in users: %v", loggedinUsers)

	return &pb.LoginResponse{Token: tokenString}, nil
}

// LogOut handles user logout requests
func (s *server) LogOut(ctx context.Context, in *pb.LogOutRequest) (*pb.LogOutResponse, error) {
	log.Printf("Received Logout request: %v", in)

	tokenString := in.AuthCode
	if tokenString == "" {
		log.Print("Missing authorization code")
		return nil, fmt.Errorf("missing authorization code")
	}

	_, err := verifyToken(tokenString)
	if err != nil {
		log.Printf("Invalid token: %v", err)
		return nil, fmt.Errorf("invalid token")
	}

	removeFromLoggedIn(in.Username)
	log.Printf("Logged out user: %s", in.Username)

	return &pb.LogOutResponse{Status: "Logged Out"}, nil
}

// Register handles user registration requests
func (s *server) Register(ctx context.Context, in *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	log.Printf("Received Register request: %v", in)

	if isUserRegistered(in.Username) {
		log.Printf("Username already registered: %s", in.Username)
		return &pb.RegisterResponse{Status: "failure"}, fmt.Errorf("username already registered")
	}

	registeredUsers[in.Username] = in.Password
	log.Printf("Registered users: %v", registeredUsers)

	return &pb.RegisterResponse{Status: "success"}, nil
}

// CreateToken generates a JWT token for a given username
func CreateToken(username string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"username": username,
		"exp":      time.Now().Add(time.Hour * 24).Unix(),
	})

	tokenString, err := token.SignedString(secretKey)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// verifyToken validates a JWT token and extracts the username
func verifyToken(tokenString string) (string, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return secretKey, nil
	})

	if err != nil {
		return "", err
	}

	if !token.Valid {
		return "", fmt.Errorf("invalid token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", fmt.Errorf("error parsing claims")
	}

	username, ok := claims["username"].(string)
	if !ok {
		return "", fmt.Errorf("username claim not found or not a string")
	}

	return username, nil
}
