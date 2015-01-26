package main

import (
	"time"
)

type MockClient struct {
	WaitFor  time.Duration
	GetValue map[string]interface{}
	GetError error
	Changes  chan string
}

func (mc *MockClient) Wait(stop chan bool) bool {
	select {
	case <-time.After(mc.WaitFor):
	case <-stop:
		return false
	}
	return true
}

func (mc *MockClient) Get(keys []string) (map[string]interface{}, error) {
	return mc.GetValue, mc.GetError
}

func (mc *MockClient) Watch(prefixes []string, changes chan string, stop chan bool) {
	mc.Changes = changes
	<-stop
}
