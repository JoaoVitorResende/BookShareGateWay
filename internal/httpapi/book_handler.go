package httpapi

import (
	"encoding/json"
	"log"
	"net/http"
)

const queueRegisterBook = "book.register"

// createBookRequest é o que o cliente PODE mandar.
// Não tem owner_id de propósito: se o cliente mandar, o Decode ignora.
type createBookRequest struct {
	Name        string `json:"name"`
	Author      string `json:"author"`
	Description string `json:"description"`
	Pages       int    `json:"pages"`
}

// registerBookMessage é o contrato com a Write API (fila book.register).
type registerBookMessage struct {
	OwnerID     string `json:"owner_id"`
	Name        string `json:"name"`
	Author      string `json:"author"`
	Description string `json:"description"`
	Pages       int    `json:"pages"`
}

type BookHandler struct {
	rpc Caller
}

func NewBookHandler(rpc Caller) *BookHandler {
	return &BookHandler{rpc: rpc}
}

// Create: POST /books (rota protegida).
func (h *BookHandler) Create(w http.ResponseWriter, r *http.Request) {
	// 1. Quem está pedindo: vem do JWT, validado pelo AuthMiddleware.
	userID, ok := UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	// 2. Lê só os campos permitidos.
	r.Body = http.MaxBytesReader(w, r.Body, maxBodyBytes)

	var req createBookRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}

	// 3. Monta a mensagem. O dono SEMPRE vem do token, nunca do cliente (proteção contra IDOR).
	body, err := json.Marshal(registerBookMessage{
		OwnerID:     userID,
		Name:        req.Name,
		Author:      req.Author,
		Description: req.Description,
		Pages:       req.Pages,
	})
	if err != nil {
		log.Printf("marshal register book: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	// 4. Envia para a Write API e devolve 201 no sucesso.
	callAndRespond(w, r.Context(), h.rpc, queueRegisterBook, body, http.StatusCreated)
}
