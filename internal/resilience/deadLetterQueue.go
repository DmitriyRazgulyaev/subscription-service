package resilience

import (
	"errors"
	"sync"
)

type DeadLetterQueue struct {
	mu       sync.Mutex
	messages []interface{}
	nextMsg  int
}

func NewDeadLetterQueue() *DeadLetterQueue {
	return &DeadLetterQueue{
		messages: make([]interface{}, 0),
		nextMsg:  0,
	}
}

func (dlq *DeadLetterQueue) Add(message interface{}) {
	dlq.mu.Lock()
	defer dlq.mu.Unlock()
	dlq.messages = append(dlq.messages, message)
}

func (dlq *DeadLetterQueue) GetAll() []interface{} {
	dlq.mu.Lock()
	defer dlq.mu.Unlock()
	return dlq.messages
}

func (dlq *DeadLetterQueue) ProcessWithDLQ(operation func(msg interface{}) error) error {
	dlq.mu.Lock()
	if dlq.nextMsg >= len(dlq.messages) {
		dlq.mu.Unlock()
		return errors.New("all messages already proceed")
	}

	err := operation(dlq.messages[dlq.nextMsg])
	if err != nil {
		return err
	}
	dlq.nextMsg++
	dlq.mu.Unlock()

	return nil
}
