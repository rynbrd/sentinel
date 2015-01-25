package main

import (
	"errors"
	"gopkg.in/BlueDragonX/go-settings.v0"
	"strings"
)

var UnsupportedClientConfig error = errors.New("unsupported client configuration")

// Configuration server client interface.
type Client interface {
	// Wait for the server to become available. The wait can be stopped by
	// sending a value to `stop` or closing it. Return true if the server came
	// online or false if the wait was canceled.
	Wait(stop chan bool) bool

	// Recursively retrieve the values for a group of `keys`. The values are
	// merged into a map of values structured as a tree.
	Get(keys []string) (map[string]interface{}, error)

	// Recursively watch each prefix in `prefixes` for changes. Send the name
	// of the changed prefix to the `changes` channel. Stop watching and exit
	// when `stop` receives `true`. Wait for the server to become available if
	// it isn't. Each failed watch attempt will be followed by an increasingly
	// longer period of sleep.
	Watch(prefixes []string, changes chan string, stop chan bool)
}

func NewClient(config *settings.Settings) (Client, error) {
	names := strings.Split(config.Key, "/")
	if len(names) == 0 {
		return nil, UnsupportedClientConfig
	}

	name := names[len(names)-1]
	switch name {
	case "etcd":
		return NewEtcdClient(config)
	default:
		return nil, UnsupportedClientConfig
	}
}
