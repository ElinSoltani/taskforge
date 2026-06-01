package model

type ConsumedMessage struct {
	Stream    string
	MessageID string
	Payload   QueueMessage
}
