package requests

import (
	"encoding/json"
	"io"
	"net/http"

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

func NewRequest[Request any](r *http.Request) (*Request, error) {

	body, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
		// TODO: wrap error
	}

	var request Request
	if err := json.Unmarshal(body, &request); err != nil {
		return nil, err
		// TODO: wrap error
	}

	return &request, nil
}
