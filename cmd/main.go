package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"slices"
	"sync"

	"github.com/pion/webrtc/v4"
	"github.com/Mawio/primavera-jet-server/internal/requests"
)

var (
	games   = make(map[string]*GameSession) // Mapping of room IDs to rooms
	gamesMu sync.Mutex
)

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

			game.clients = slices.DeleteFunc(game.clients, func(c *Client) bool {
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

	game := getOrCreateGame(request.SessionID)
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
	http.HandleFunc("/create", corsHandler(createSessionHandler))
	http.HandleFunc("/join", corsHandler(joinSessionHandler))

	// Start the server
	fmt.Println("Server started at :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
