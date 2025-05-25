package main

import (
	"errors"

	"github.com/pion/webrtc/v4"
)

type Connection struct {
	peerConnection *webrtc.PeerConnection
	dataChannel    *webrtc.DataChannel
}

type CreateSessionRequest struct {
	Username string                    `json:"username"`
	Offer    webrtc.SessionDescription `json:"offer"`
}

type JoinSessionRequest struct {
	Username string                    `json:"username"`
	Offer    webrtc.SessionDescription `json:"offer"`
	GameID   string                    `json:"game_id"`
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

