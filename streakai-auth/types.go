package main

import pb "streakauth/grpc"

type server struct {
	pb.UnimplementedStreakAiServiceServer
}

var secretKey = []byte("secret-key")
var registeredUsers = map[string]string{}
var loggedinUsers = []string{}
