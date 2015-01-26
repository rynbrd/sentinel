package main

import (
	"gopkg.in/BlueDragonX/go-settings.v0"
	"testing"
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

func TestNewClient(t *testing.T) {
	config := &settings.Settings{}
	config.Set("etcd.uris", []string{"http://localhost:4001/"})
	config.Set("consul.uris", []string{"http://localhost:4001/"})

	clientConfig, _ := config.Object("etcd")
	if client, err := NewClient(clientConfig); err == nil {
		if _, ok := client.(*EtcdClient); !ok {
			t.Error("etcd client not returned for etcd config")
		}
	} else {
		t.Error(err)
	}

	clientConfig, _ = config.Object("consul")
	if _, err := NewClient(clientConfig); err != UnsupportedClientConfig {
		t.Error("NewClient did not return a client error")
	}
}
