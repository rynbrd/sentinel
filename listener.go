package main

import (
	"github.com/coreos/go-etcd/etcd"
	"gopkg.in/BlueDragonX/simplelog.v1"
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
	logger *simplelog.Logger
	stop   chan bool
	join   chan bool
}

// Create a new watcher. The watcher immediately begins monitoring etcd for changes.
func NewListener(prefix, key string, client *Client, logger *simplelog.Logger) *Listener {
	return &Listener{
		key,
		prefix,
		client,
		logger,
		make(chan bool),
		make(chan bool),
	}
}

// Start the listener. Emit the name of the key to the provided channel when it changes.
func (w *Listener) Start(events []chan string) {
	key := joinPaths(w.prefix, w.Key)

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
					event := strings.Trim(strings.TrimPrefix(response.Node.Key, w.prefix), "/")
					for _, eventChan := range events {
						eventChan <- event
					}
				}
				join <- true
				close(join)
			}()

			_, err := w.client.client.Watch(key, 0, false, responses, w.stop)
			<-join

			if err == etcd.ErrWatchStoppedByUser {
				break Loop
			} else {
				w.logger.Error("watch on %s failed: %s", key, err)
				w.logger.Info("retrying in %ds", WatchRetry)
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
