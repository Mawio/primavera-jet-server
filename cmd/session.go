package main

import (
	"fmt"
	"slices"
	"sync"

	"github.com/pion/webrtc/v4"
)

type Session struct {
	id      string
	users []*User
	mu      sync.Mutex
}

var (
	sessionsMu sync.Mutex
)

func getOrCreateSession(sessionId string) *Session {
	sessionsMu.Lock()
	defer sessionsMu.Unlock()

	if session, ok := sessions[sessionId]; ok {
		return session
	}

	session := &Session{}
	sessions[sessionId] = session
	return session
}

func (session *Session) addUser(user *User) {
	user.connection.peerConnection.OnDataChannel(func(dataChannel *webrtc.DataChannel) {
		fmt.Printf("New DataChannel for user %s\n", user.username)

		dataChannel.OnOpen(func() {
			fmt.Printf("Data channel with user %s open.\n", user.username)
		})

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

	session.users = append(session.users, user)
}

