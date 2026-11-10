package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"
)

const rpcTimeout = 5 * time.Second

// callAndRespond envia o body para a fila, espera a resposta
// e traduz o envelope para uma resposta HTTP.
func callAndRespond(w http.ResponseWriter, ctx context.Context, rpc Caller, queue string, body []byte, successStatus int) {
	ctx, cancel := context.WithTimeout(ctx, rpcTimeout) // ajuste 1: r.Context() vira o parâmetro ctx
	defer cancel()

	resp, err := rpc.Call(ctx, queue, body) // ajuste 2: h.rpc vira o parâmetro rpc
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
	if env.Error == nil { // ajuste 3: sucesso usa o status que o handler pediu
		status = successStatus
	}

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