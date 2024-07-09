package main

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/go-redis/redis"
)

func initRedis() {
	redisClient = redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})

	_, err := redisClient.Ping().Result()
	if err != nil {
		log.Fatalf("Error connecting to Redis: %v", err)
	}
}

func getSession(sessionID string) (*VotingSession, error) {
	sessionData, err := redisClient.Get(sessionID).Result()
	if err == redis.Nil {
		return nil, fmt.Errorf("session not found")
	} else if err != nil {
		return nil, fmt.Errorf("failed to get session from Redis: %v", err)
	}

	var session VotingSession
	if err := json.Unmarshal([]byte(sessionData), &session); err != nil {
		return nil, fmt.Errorf("failed to unmarshal session data: %v", err)
	}

	return &session, nil
}

func setSession(session *VotingSession) error {
	sessionMutex.Lock()
	defer sessionMutex.Unlock()
	sessions = append(sessions, session)
	sessionData, err := json.Marshal(session)
	if err != nil {
		return fmt.Errorf("failed to marshal session: %v", err)
	}

	if err := redisClient.Set(session.Id, sessionData, 0).Err(); err != nil {
		return fmt.Errorf("failed to set session in Redis: %v", err)
	}

	return nil
}
