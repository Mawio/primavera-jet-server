package main

import (
	"sync"

	"github.com/pion/webrtc/v4"
)

type Connection struct {
	peerConnection *webrtc.PeerConnection
	dataChannel    *webrtc.DataChannel
}

type Client struct {
	username string
	connection Connection
}

type GameSession struct {
	id      string
	clients []*Client
	mu      sync.Mutex
}

