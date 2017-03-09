package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"testing"
	"time"
)

func TestSimpleExample(t *testing.T) {
	// Pick a port for the http status server
	port := getRandomFreePort()

	os.Args = []string{"example", "-addr", fmt.Sprintf(":%v", port)}
	go main()
	<-ready

	// Block so the main starts
	time.Sleep(100 * time.Millisecond)

	// Get the status endpoint
	res, err := http.Get(fmt.Sprintf("http://127.0.0.1:%v", port))
	if err != nil {
		t.Fatal(err)
	}
	status, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		t.Fatalf("Read request body: %v", err)
	}

	s := &struct {
		Name    string `json:"name"`
		Group   string `json:"group"`
		Version string `json:"version"`
	}{}

	if err := json.Unmarshal(status, s); err != nil {
		t.Fatalf("Unmarshaling response: Unexpected error %v", err)
	}

	if s.Group != "example" || s.Name != "example" || s.Version != "1.0" {
		t.Errorf("Unexpected status response: %v", s)
	}

	hires, err := service.Request("example.hi", nil, timeout)
	if err != nil {
		t.Errorf("Requesting example.hi.proxy: Unexpected error %v", err)
	}

	if string(hires) != "Hello world" {
		t.Errorf("Unexpected response on example.hi.proxy: %v", string(hires))
	}

}

func getRandomFreePort() int {
	l, _ := net.Listen("tcp", ":0")
	err := l.Close()
	if err != nil {
		panic(err)
	}

	return l.Addr().(*net.TCPAddr).Port
}
