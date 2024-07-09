package main

import (
	"net/http"
	pb "streakai/grpc"
	"sync"

	"github.com/go-redis/redis"
	"github.com/gorilla/websocket"
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
	wsMutex     sync.Mutex
	clients     = make(map[*websocket.Conn]bool)
	redisClient *redis.Client
)
