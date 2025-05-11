package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"slices"
	"sync"

	"github.com/pion/webrtc/v4"
)

var (
	games   = make(map[string]*GameSession) // Mapping of room IDs to rooms
	gamesMu sync.Mutex
)

// Generates a unique room ID.
func generateGameID() string {
	//TODO
	return "primavera"
}

func setupWebRtcConnection(offer webrtc.SessionDescription) (Connection, error) {
	// Create a new PeerConnection
	peerConnection, err := webrtc.NewPeerConnection(webrtc.Configuration{})
	if err != nil {
		return Connection{}, errors.New("Failed to create peer connection")
	}

	// Set the remote SDP offer
	err = peerConnection.SetRemoteDescription(offer)
	if err != nil {
		return Connection{}, errors.New("Failed to set remote description")
	}

	// Create an SDP answer
	answer, err := peerConnection.CreateAnswer(nil)
	if err != nil {
		return Connection{}, errors.New("Failed to create answer")
	}

	// Set local SDP answer
	err = peerConnection.SetLocalDescription(answer)
	if err != nil {
		return Connection{}, errors.New("Failed to set local description")
	}

	// Wait until ICE gathering is complete
	gatherComplete := webrtc.GatheringCompletePromise(peerConnection)
	<-gatherComplete

	return Connection{
		peerConnection: peerConnection,
	}, nil
}

func setupHandlers(client *Client, game *GameSession) {
	client.connection.peerConnection.OnDataChannel(func(dataChannel *webrtc.DataChannel) {
		fmt.Printf("New DataChannel for client %s\n", client.username)

		// Register channel opening handling
		dataChannel.OnOpen(func() {
			fmt.Printf("Data channel with client %s open.\n", client.username)
		})

		// Register text message handling
		dataChannel.OnMessage(func(msg webrtc.DataChannelMessage) {
			fmt.Printf("Received message from %s\n", client.username)
			// Forward the message to all other peers in the game
			game.mu.Lock()
			defer game.mu.Unlock()

			for _, c := range game.clients {
				if c.username != client.username && c.connection.dataChannel != nil {
					fmt.Printf("Sending message to %s\n", c.username)
					_ = c.connection.dataChannel.Send(msg.Data)
				}
			}
		})

		client.connection.dataChannel = dataChannel
	})

	// Handle peer connection state changes
	client.connection.peerConnection.OnConnectionStateChange(func(state webrtc.PeerConnectionState) {
		if state == webrtc.PeerConnectionStateDisconnected || state == webrtc.PeerConnectionStateFailed || state == webrtc.PeerConnectionStateClosed {
			game.mu.Lock()
			defer game.mu.Unlock()

			slices.DeleteFunc(game.clients, func(c *Client) bool {
				return c == client
			})
		}
	})
}

// Get or create a room by its ID.
func getOrCreateGame(gameID string) *GameSession {
	gamesMu.Lock()
	defer gamesMu.Unlock()

	// If the game exists, return it. Otherwise, create a new one.
	if game, ok := games[gameID]; ok {
		return game
	}

	// Create a new game if it doesn't exist
	game := &GameSession{}
	games[gameID] = game
	return game
}

// /create - Create a new game and return a unique ID.
func createGameHandler(w http.ResponseWriter, r *http.Request) {
	// Generate a new unique game ID
	var request CreateGameRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	connection, err := setupWebRtcConnection(request.Offer)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	gameID := generateGameID()
	game := getOrCreateGame(gameID)

	client := &Client{
		username:   request.Username,
		connection: connection,
	}

	setupHandlers(client, game)

	game.clients = append(game.clients, client)

	// Respond with the room ID
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"game_id": gameID,
		"SDP":     connection.peerConnection.LocalDescription(), // Assuming connection is the peer connection
	})
}

// /join - Join an existing room.
func joinRoomHandler(w http.ResponseWriter, r *http.Request) {
	var request JoinGameRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	connection, err := setupWebRtcConnection(request.Offer)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	game := getOrCreateGame(request.GameID)
	game.mu.Lock()
	defer game.mu.Unlock()

	if slices.ContainsFunc(game.clients, func(client *Client) bool {
		return client.username == request.Username
	}) {
		http.Error(w, "Username already in use in this game", http.StatusInternalServerError)
		return
	}

	client := &Client{
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
	http.HandleFunc("/create", corsHandler(createGameHandler))
	http.HandleFunc("/join", corsHandler(joinRoomHandler))

	// Start the server
	fmt.Println("Server started at :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
