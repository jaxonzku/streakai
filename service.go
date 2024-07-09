package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

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
	username, _ := isAuthorised(w, r)
	for i, session := range sessions {
		log.Printf("%d: %+v", i, session)
		if session.Id == singleVote.Id {
			if !alreadyVoted(session.YesCount, session.NoCount, username) {
				if singleVote.Vote {
					session.YesCount = append(session.YesCount, username)
					broadcastSessionStatus(session)
					SendResponse(w, http.StatusOK, map[string]string{"message": "vote casted"})
				} else {
					session.NoCount = append(session.NoCount, username)
					broadcastSessionStatus(session)
					SendResponse(w, http.StatusOK, map[string]string{"message": "vote casted"})

				}
			}

		}
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
		sessionMutex.Lock()
		sessions = append(sessions, &votingSession)
		sessionMutex.Unlock()
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

	fmt.Println("yes", yesCount)
	fmt.Println("no", noCount)

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
	fmt.Println("a1", yes || no)
	fmt.Println("a2", yes && no)
	fmt.Println("a3", yes)
	fmt.Println("a3", no)

	return yes || no
}
