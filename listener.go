package main

import (
	"github.com/coreos/go-etcd/etcd"
	"strings"
	"time"
)

const (
	WatchRetry = 5
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
	logger.Debug("watching '%s'", key)

	go func() {
	Loop:
		for {
			join := make(chan bool)
			responses := make(chan *etcd.Response)
			go func() {
				for {
					response, open := <-responses
					if !open {
						break
					}
					logger.Debug("key '%s' changed", response.Node.Key)
					event := strings.Trim(strings.TrimPrefix(response.Node.Key, w.prefix), "/")
					for _, eventChan := range events {
						eventChan <- event
					}
				}
				join <- true
				close(join)
			}()

			_, err := w.client.client.Watch(key, 0, true, responses, w.stop)
			<-join

			if err == etcd.ErrWatchStoppedByUser {
				break Loop
			} else {
				logger.Error("watch on '%s' failed: %s", key, err)
				logger.Info("retrying in %ds", WatchRetry)
				select {
				case <-w.stop:
					break Loop
				case <-time.After(WatchRetry * time.Second):
				}
			}
		}
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
