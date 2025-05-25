package requests

import (
	"github.com/pion/webrtc/v4"
)

type CreateSession struct {
	Username string                    `json:"username"`
	Offer    webrtc.SessionDescription `json:"offer"`
}

type JoinSession struct {
	Username  string                    `json:"username"`
	Offer     webrtc.SessionDescription `json:"offer"`
	SessionID string                    `json:"session_id"`
}
