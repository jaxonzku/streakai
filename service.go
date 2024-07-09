package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
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
	username, _ := isAuthorised(w, r)
	for i, session := range sessions {
		log.Printf("%d: %+v", i, session)
		if session.Id == singleVote.Id {
			if !alreadyVoted(session.YesCount, session.NoCount, username) {
				if singleVote.Vote {
					session.YesCount = append(session.YesCount, username)
					setSession(session)
					broadcastSessionStatus(session)
					SendResponse(w, http.StatusOK, map[string]string{"message": "vote casted"})
				} else {
					session.NoCount = append(session.NoCount, username)
					setSession(session)
					broadcastSessionStatus(session)
					SendResponse(w, http.StatusOK, map[string]string{"message": "vote casted"})

				}
			}

		}
	}

}

func getSessionHandler(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	sessionID, ok := vars["id"]
	if !ok {
		http.Error(w, "Missing session ID", http.StatusBadRequest)
		return
	}
	session, err := getSession(sessionID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(session); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

}

func createSessionHandler(w http.ResponseWriter, r *http.Request) {
	_, isAuthorised := isAuthorised(w, r)
	if isAuthorised {
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
		setSession(&votingSession)
		broadcastSessionStatus(&votingSession)
		logAllSessions()
	}

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

func alreadyVoted(yesCount []string, noCount []string, username string) bool {
	yes := false
	no := false

	for _, u := range yesCount {
		fmt.Println("user yes", u)
		if u == username {
			yes = true
			break
		}
	}

	for _, u := range noCount {
		fmt.Println("user no", u)
		if u == username {
			no = true
			break
		}
	}

	return yes || no
}
