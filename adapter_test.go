package mrpc

import "testing"

func TestEmptyMessageAdapter(t *testing.T) {
	a := emptyMessageAdapter{}
	err := a.ProcessMessage(reqMessageType, "topicName", []byte{})
	if err != nil {
		t.Errorf("Unexpected error")
	}
}
