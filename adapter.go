package mrpc

const (
	reqMessageType MessageType = iota
	resMessageType
	pubMessageType
	subMessageType
)

// MessageType is the type of the message
type MessageType int

// MessageAdapter is used to intercept and process messages
type MessageAdapter interface {
	ProcessMessage(t MessageType, topicName string, data []byte) error
}

type emptyMessageAdapter struct{}

func (a *emptyMessageAdapter) ProcessMessage(t MessageType, topicName string, data []byte) error {
	return nil
}
