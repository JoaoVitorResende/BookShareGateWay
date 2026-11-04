package messaging

import (
	"context"
	"sync"

	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"
)

type RPCClient struct {
	ch         *amqp.Channel
	replyQueue string
	mu         sync.Mutex
	pending    map[string]chan []byte
}


func NewRPCClient(ch *amqp.Channel) (*RPCClient, error) {
	// fila anônima, exclusiva e auto-delete: só este gateway lê
	q, err := ch.QueueDeclare("", false, true, true, false, nil)
	if err != nil {
		return nil, err
	}
	msgs, err := ch.Consume(q.Name, "", true, true, false, false, nil)
	if err != nil {
		return nil, err
	}

	c := &RPCClient{ch: ch, replyQueue: q.Name, pending: make(map[string]chan []byte)}
	go c.listen(msgs)
	return c, nil
}

func (c *RPCClient) listen(msgs <-chan amqp.Delivery) {
	for d := range msgs {
		c.mu.Lock()
		waiter, ok := c.pending[d.CorrelationId]
		c.mu.Unlock()
		if ok { // se não achou, o cliente já deu timeout: descarta
			waiter <- d.Body
		}
	}
}

func (c *RPCClient) Call(ctx context.Context, queue string, body []byte) ([]byte, error) {
	id := uuid.NewString()
	waiter := make(chan []byte, 1) // buffer 1: listen nunca bloqueia

	c.mu.Lock()
	c.pending[id] = waiter
	c.mu.Unlock()
	defer func() {
		c.mu.Lock()
		delete(c.pending, id)
		c.mu.Unlock()
	}()

	err := c.ch.PublishWithContext(ctx, "", queue, false, false, amqp.Publishing{
		ContentType:   "application/json",
		CorrelationId: id,
		ReplyTo:       c.replyQueue,
		Body:          body,
	})
	if err != nil {
		return nil, err
	}

	select {
	case resp := <-waiter:
		return resp, nil
	case <-ctx.Done():
		return nil, ctx.Err() // timeout
	}
}
