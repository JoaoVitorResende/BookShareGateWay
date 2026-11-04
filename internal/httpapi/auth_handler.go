package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"time"
)

type AuthHandler struct {
	rpc Caller
}

func NewAuthHandler(rpc Caller) *AuthHandler {
	return &AuthHandler{rpc: rpc}
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	h.forward(w, r, "user.register")
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	h.forward(w, r, "user.login")
}

func (h *AuthHandler) forward(w http.ResponseWriter, r *http.Request, queue string) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	body, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
	if err != nil || !json.Valid(body) {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}

	resp, err := h.rpc.Call(ctx, queue, body)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			http.Error(w, "upstream timeout", http.StatusGatewayTimeout)
			return
		}
		http.Error(w, "upstream error", http.StatusBadGateway)
		return
	}

	var env replyEnvelope
	if err := json.Unmarshal(resp, &env); err != nil {
		http.Error(w, "invalid upstream response", http.StatusBadGateway)
		return
	}

	status := statusFromEnvelope(env)

	var payload []byte
	if env.Error != nil {
		payload, _ = json.Marshal(map[string]*errorBody{"error": env.Error})
	} else if len(env.Data) > 0 {
		payload = env.Data
	} else {
		payload = []byte("null")
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(payload)
}