package main

import (
	"fmt"

	"github.com/pion/webrtc/v4"
)

type User struct {
	username   string
	connection Connection
}

func NewUser(offer webrtc.SessionDescription, username string) *User {
	connection, err := setupWebRtcConnection(offer)
	if err != nil {
		return nil
	}

	return &User{
		username: username,
		connection: connection,
	}
}

func (user User) sendMessage(data []byte) {
	if user.connection.dataChannel != nil {
		fmt.Printf("Sending message to %s\n", user.username)
		_ = user.connection.dataChannel.Send(data)
	}
}
