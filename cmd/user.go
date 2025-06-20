package main

import (
	"fmt"
	"slices"
)

type User struct {
	username string
	connection Connection
}

func (user *User) setupHandlers(session *Session) {
	user.connection.peerConnection.OnDataChannel(func(dataChannel *webrtc.DataChannel) {
		fmt.Printf("New DataChannel for client %s\n", user.username)

		// Register channel opening handling
		dataChannel.OnOpen(func() {
			fmt.Printf("Data channel with client %s open.\n", user.username)
		})

		// Register text message handling
		dataChannel.OnMessage(func(msg webrtc.DataChannelMessage) {
			fmt.Printf("Received message from %s\n", user.username)
			// Forward the message to all other peers in the game
			session.mu.Lock()
			defer session.mu.Unlock()

			for _, u := range session.users {
				if u.username != user.username && u.connection.dataChannel != nil {
					fmt.Printf("Sending message to %s\n", u.username)
					_ = u.connection.dataChannel.Send(msg.Data)
				}
			}
		})

		user.connection.dataChannel = dataChannel
	})

	// Handle peer connection state changes
	user.connection.peerConnection.OnConnectionStateChange(func(state webrtc.PeerConnectionState) {
		if state == webrtc.PeerConnectionStateDisconnected || state == webrtc.PeerConnectionStateFailed || state == webrtc.PeerConnectionStateClosed {
			session.mu.Lock()
			defer session.mu.Unlock()

			session.users = slices.DeleteFunc(session.users, func(c *User) bool {
				return c == user
			})
		}
	})
}

