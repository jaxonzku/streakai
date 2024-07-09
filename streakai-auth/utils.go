package main

import (
	"context"
	"fmt"
	"log"
	pb "streakauth/grpc"
)

// CheckAuthorized handles authorization check requests
func (s *server) CheckAuthorized(ctx context.Context, in *pb.CheckAuthorizedReq) (*pb.CheckAuthorizedRes, error) {
	log.Printf("Received authorization check request: %v", in)

	username, err := verifyToken(in.AuthCode)
	if err != nil {
		log.Printf("Invalid token: %v", err)
		return &pb.CheckAuthorizedRes{Username: "", Authorized: false}, fmt.Errorf("invalid token")
	}

	return &pb.CheckAuthorizedRes{Username: username, Authorized: true}, nil
}

// isUserRegistered checks if a user is already registered
func isUserRegistered(username string) bool {
	_, exists := registeredUsers[username]
	return exists
}

// removeFromLoggedIn removes a user from the logged-in users list
func removeFromLoggedIn(username string) {
	for i, u := range loggedinUsers {
		if u == username {
			loggedinUsers = append(loggedinUsers[:i], loggedinUsers[i+1:]...)
			break
		}
	}
}
