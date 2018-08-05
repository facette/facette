package v1

import "fmt"

// Message represents an API message instance.
type Message struct {
	Message string `json:"message"`
}

func newMessage(err error) *Message {
	return &Message{fmt.Sprintf("%s", err)}
}
