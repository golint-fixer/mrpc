package mrpc

import (
	"errors"
	"fmt"
	"testing"
)

var errFakeAdapter = errors.New("fake adapter error")

type fakeAdapter struct{}

func (fakeAdapter) ProcessMessage(t MessageType, topicName string, data []byte) error {
	return errFakeAdapter
}

func TestSimpleFuncOptions(t *testing.T) {
	cases := []struct {
		f func(*Service) error
		t func(*Service) bool
	}{
		{
			f: WithName("testName"),
			t: func(s *Service) bool {
				return s.name == "testName"
			},
		},
		{
			f: WithGroup("testGroup"),
			t: func(s *Service) bool {
				return s.group == "testGroup"
			},
		},
		{
			f: WithVersion("testVer"),
			t: func(s *Service) bool {
				return s.version == "testVer"
			},
		},
		{
			f: WithNGV("testName", "testGroup", "testVer"),
			t: func(s *Service) bool {
				return s.name == "testName" && s.group == "testGroup" && s.version == "testVer"
			},
		},
		{
			f: WithAdapter(fakeAdapter{}),
			t: func(s *Service) bool {
				return s.adapter.ProcessMessage(reqMessageType, "", nil) == errFakeAdapter
			},
		},
	}

	for i, tc := range cases {
		t.Run(fmt.Sprintf("Case%v", i), func(t *testing.T) {
			s, _ := NewService(&FakeTransport{}, tc.f)

			if !tc.t(s) {
				t.Error("is not giving the expected outcome")
			}
		})
	}
}
