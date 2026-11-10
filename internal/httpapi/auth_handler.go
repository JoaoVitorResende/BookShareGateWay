package httpapi

import (
	"encoding/json"
	"io"
	"net/http"
)

const maxBodyBytes = 1 << 20 // 1 MB

type AuthHandler struct {
	rpc Caller
}

func NewAuthHandler(rpc Caller) *AuthHandler {
	return &AuthHandler{rpc: rpc}
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	h.forward(w, r, "user.register", http.StatusCreated)
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	h.forward(w, r, "user.login", http.StatusOK)
}

// forward repassa o JSON do cliente para a fila sem alterar nada.
func (h *AuthHandler) forward(w http.ResponseWriter, r *http.Request, queue string, successStatus int) {
	body, err := io.ReadAll(io.LimitReader(r.Body, maxBodyBytes))
	if err != nil || !json.Valid(body) {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}

	callAndRespond(w, r.Context(), h.rpc, queue, body, successStatus)
}
