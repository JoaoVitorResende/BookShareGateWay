package httpapi

import (
	"encoding/json"
	"net/http"
)

type errorBody struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type replyEnvelope struct {
	OK    bool            `json:"ok"`
	Data  json.RawMessage `json:"data"`
	Error *errorBody      `json:"error"`
}

// statusFromEnvelope traduz o código de erro do contrato da fila para status HTTP.
func statusFromEnvelope(env replyEnvelope) int {
	if env.Error == nil {
		return http.StatusOK
	}
	switch env.Error.Code {
	case "EMAIL_ALREADY_EXISTS", "BOOK_ALREADY_EXISTS":
		return http.StatusConflict
	case "INVALID_CREDENTIALS", "UNAUTHORIZED":
		return http.StatusUnauthorized
	case "INVALID_INPUT":
		return http.StatusBadRequest
	default:
		return http.StatusInternalServerError
	}
}
