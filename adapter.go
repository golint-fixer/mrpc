package mrpc

const (
	MESSAGETYPE_REQUEST = iota
	MESSAGETYPE_RESPONSE
	MESSAGETYPE_PUBLISH
	MESSAGETYPE_SUBSCRIBE
)

type MessageAdapter interface {
	SetServiceInfo(serviceName, serviceGroup, serviceVersion string)
	ProcessMessage(messageType int, topicName string, data []byte) error
}

type EmptyMessageAdapter struct{}

func (a *EmptyMessageAdapter) SetServiceInfo(serviceName, serviceGroup, serviceVersion string) {}

func (a *EmptyMessageAdapter) ProcessMessage(messageType int, topicName string, data []byte) error {
	return nil
}
