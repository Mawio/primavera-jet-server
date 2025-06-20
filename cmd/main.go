package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"slices"

	"github.com/Mawio/primavera-jet-server/internal/requests"
)

var (
	sessions   = make(map[string]*Session)
)

// /create - Create a new session and return a unique ID.
func createSessionHandler(w http.ResponseWriter, r *http.Request) {
	request, err := requests.DecodeRequest[requests.CreateSession](r)
	if err != nil {
		//TODO: handle error
		return
	}

	connection, err := setupWebRtcConnection(request.Offer)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	session := getOrCreateSession(generateSessionId())

	session.addUser(&User{
		username:   request.Username,
		connection: connection,
	})

	// Respond with the sessionId 
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"game_id": session.id,
		"SDP":     connection.peerConnection.LocalDescription(), // Assuming connection is the peer connection
	})
}

// /join - Join an existing session.
func joinSessionHandler(w http.ResponseWriter, r *http.Request) {
	request, err := requests.DecodeRequest[requests.JoinSession](r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	connection, err := setupWebRtcConnection(request.Offer)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	game := getOrCreateSession(request.SessionID)
	game.mu.Lock()
	defer game.mu.Unlock()

	if slices.ContainsFunc(game.clients, func(client *User) bool {
		return client.username == request.Username
	}) {
		http.Error(w, "Username already in use in this game", http.StatusInternalServerError)
		return
	}

	client := &User{
		username:   request.Username,
		connection: connection,
	}

	setupHandlers(client, game)

	game.clients = append(game.clients, client)

	// Return the answer SDP to the client
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"SDP": connection.peerConnection.LocalDescription(),
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
