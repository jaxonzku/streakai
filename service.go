package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	pb "streakai/grpc"

	"github.com/google/uuid"
)

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
