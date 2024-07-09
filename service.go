package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

// castVote handles the voting process for a session
func castVote(w http.ResponseWriter, r *http.Request) {
	log.Println("Handling cast vote request")
	defer r.Body.Close()

	var singleVote SingleVote
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
				} else {
					session.NoCount = append(session.NoCount, username)
				}
				setSession(session)
				broadcastSessionStatus(session)
				SendResponse(w, http.StatusOK, map[string]string{"message": "vote cast"})
				return
			} else {
				http.Error(w, "User has already voted", http.StatusConflict)
				return
			}
		}
	}
	http.Error(w, "Session not found", http.StatusNotFound)
}

// getSessionHandler handles fetching a session by ID
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

// createSessionHandler handles the creation of a new voting session
func createSessionHandler(w http.ResponseWriter, r *http.Request) {
	_, isAuthorised := isAuthorised(w, r)
	if isAuthorised {
		log.Println("Handling create voting session request")
		defer r.Body.Close()

		var votingSession VotingSession
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
		SendResponse(w, http.StatusOK, map[string]string{"message": "Session created", "sessionID": votingSession.Id})
	} else {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
	}
}

// SendResponse sends a JSON response with a given status code and payload
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

// alreadyVoted checks if a user has already voted in the session
func alreadyVoted(yesCount []string, noCount []string, username string) bool {
	yes := false
	no := false

	for _, u := range yesCount {

		if u == username {
			yes = true
			break
		}
	}

	for _, u := range noCount {

		if u == username {
			no = true
			break
		}
	}

	return yes || no
}
