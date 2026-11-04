package httpapi

import "encoding/json"

type replyEnvelope struct {
	Status int             `json:"status"`
	Data   json.RawMessage `json:"data"`
	Error  string          `json:"error"`
}