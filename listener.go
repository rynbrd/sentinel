package main

import (
	"time"
)

// A Listener waits for etcd key changes and sends watch events to its eventss.
type Listener struct {
	Key    string
	prefix string
	client *Client
	stop   chan bool
	join   chan bool
}

// Create a new listener. The listener immediately begins monitoring etcd for changes.
func NewListener(prefix, key string, client *Client) *Listener {
	return &Listener{
		key,
		prefix,
		client,
		make(chan bool),
		make(chan bool),
	}
}

// Start the listener. Emit the name of the key to the provided channel when it changes.
func (w *Listener) Start(events []chan string) {
	key := joinPaths(w.prefix, w.Key)
	logger.Debugf("watching '%s'", key)

	go func() {
		join := make(chan struct{})
		changes := make(chan string)
		go func() {
			for {
				change, open := <-changes
				if !open {
					break
				}
				for _, eventChan := range events {
					eventChan <- change
				}
			}
			close(join)
		}()

		w.client.Watch([]string{key}, changes, w.stop)
		<-join

		for _, eventChan := range events {
			close(eventChan)
		}
		w.join <- true
	}()
}

// Stop a Listener.
func (w *Listener) Stop() {
	w.stop <- true
	select {
	case <-w.join:
	case <-time.After(200 * time.Millisecond):
	}
}
