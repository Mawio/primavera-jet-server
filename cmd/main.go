package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/Mawio/primavera-jet-server/internal/requests"
)

// /create - Create a new session and return a unique ID.
func createSessionHandler(w http.ResponseWriter, r *http.Request) {
	request, err := requests.DecodeRequest[requests.CreateSession](r)
	if err != nil {
		//TODO: handle error
		return
	}

	session, err := createSession(generateSessionId())
	if err != nil {
		//TODO: handle error
		return
	}

	user := NewUser(request.Offer, request.Username)
	session.addUser(user)

	// Respond with the sessionId 
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"game_id": session.id,
		"SDP":     user.connection.peerConnection.LocalDescription(), // Assuming connection is the peer connection
	})
}

// /join - Join an existing session.
func joinSessionHandler(w http.ResponseWriter, r *http.Request) {
	request, err := requests.DecodeRequest[requests.JoinSession](r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	session, err := getSession(request.SessionID)
	if err != nil {
		//TODO: handle error
		return
	}

	user := NewUser(request.Offer, request.Username)
	session.addUser(user)

	// Return the answer SDP to the client
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"SDP": user.connection.peerConnection.LocalDescription(),
	})
}

func main() {
	// Set up HTTP routes
	http.HandleFunc("/create", corsHandler(createSessionHandler))
	http.HandleFunc("/join", corsHandler(joinSessionHandler))

	// Start the server
	fmt.Println("Server started at :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
