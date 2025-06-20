package main

import (
	"fmt"
	"slices"
	"sync"

	"github.com/pion/webrtc/v4"
)

type Session struct {
	id    string
	users []*User
	mu    sync.Mutex
}

var (
	sessions   = make(map[string]*Session)
	sessionsMu sync.Mutex
)

func createSession(sessionId string) (*Session, error) {
	sessionsMu.Lock()
	defer sessionsMu.Unlock()

	if _, ok := sessions[sessionId]; ok {
		return nil, fmt.Errorf("Session %s already exists", sessionId)
	}

	session := &Session{}
	sessions[sessionId] = session
	return session, nil
}

func getSession(sessionId string) (*Session, error) {
	sessionsMu.Lock()
	defer sessionsMu.Unlock()

	if session, ok := sessions[sessionId]; ok {
		return session, nil
	} else {
		return nil, fmt.Errorf("Session %s does not exist", sessionId)
	}

}

func (session *Session) addUser(user *User) error {

	if slices.ContainsFunc(session.users, func(u *User) bool {
		return user.username == u.username
	}) {
		return fmt.Errorf("Username already in use in this session")
	}

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
				if u.username != user.username{
					u.sendMessage(msg.Data)
				}}
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
	return nil
}
