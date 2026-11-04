package httpapi

import "context"

type Caller interface {
	Call(ctx context.Context, queue string, body []byte) ([]byte, error)
}