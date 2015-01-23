package main

import (
	"errors"
	"fmt"
	"gopkg.in/BlueDragonX/go-settings.v0"
	"strings"
)

// A watcher renders templates and executes command when key changes are detected.
type Watcher struct {
	name     string
	prefix   string
	watch    []string
	context  []string
	executor *Executor
	client   *Client
}

// Create a new watcher with the provided configuration.
func NewWatcher(client *Client, config *settings.Settings) (*Watcher, error) {
	names := strings.Split(config.Key, ".")
	name := names[len(names)-1]
	prefix := config.StringDflt("prefix", client.prefix)
	watch := config.StringArrayDflt("watch", []string{})
	context := config.StringArrayDflt("context", []string{})

	if len(watch) == 0 {
		return nil, fmt.Errorf("watcher %s requires at least one watch path", name)
	}
	if len(context) == 0 {
		return nil, fmt.Errorf("watcher %s requires at least one context path", name)
	}

	var command []string
	if cmdStr, err := config.String("command"); err == nil {
		command = []string{"bash", "-c", cmdStr}
	} else if cmdArray, err := config.StringArray("command"); err == nil {
		command = cmdArray
	}

	tplConfigs := config.ObjectArrayDflt("templates", []*settings.Settings{})
	tpls := make([]Template, len(tplConfigs))
	for n, tplConfig := range tplConfigs {
		tpl := Template{
			Src:  tplConfig.StringDflt("src", ""),
			Dest: tplConfig.StringDflt("dest", ""),
		}
		if tpl.Src == "" {
			return nil, fmt.Errorf("watcher %s template %d requires a src path", name, n)
		}
		if tpl.Dest == "" {
			return nil, fmt.Errorf("watcher %s template %d requires a dest path", name, n)
		}
		tpls = append(tpls, tpl)
	}

	executor := &Executor{
		Name:      name,
		Templates: tpls,
		Command:   command,
	}

	return &Watcher{
		name:     name,
		prefix:   prefix,
		watch:    watch,
		context:  context,
		executor: executor,
		client:   client,
	}, nil
}

// Return the name of the watcher.
func (watcher *Watcher) Name() string {
	return watcher.name
}

// Execute the watcher as if an event was receieved.
func (watcher *Watcher) Execute() error {
	context, err := watcher.client.Get(watcher.prefix, watcher.context)
	if err != nil {
		logger.Errorf("%s failed to retrieve context: %s", watcher.Name(), err)
		return err
	}
	return watcher.executor.Execute(context)
}

// Execute the watcher when an event is received.
func (watcher *Watcher) Run(events chan string) {
	for {
		_, open := <-events
		if !open {
			break
		}
		watcher.Execute()
	}
}

// Manage multiple watchers.
type WatchManager struct {
	Watchers  map[string]*Watcher
	listeners map[string]*Listener
	client    *Client
}

// Create a new watch manager.
func NewWatchManager(client *Client, watchers []*Watcher) *WatchManager {
	manager := &WatchManager{
		make(map[string]*Watcher),
		make(map[string]*Listener),
		client,
	}

	for _, watcher := range watchers {
		manager.Watchers[watcher.Name()] = watcher
		for _, key := range watcher.watch {
			if _, have := manager.listeners[key]; !have {
				manager.listeners[key] = NewListener(watcher.prefix, key, client)
			}
		}
	}
	return manager
}

// Execute the named watchers.
func (manager *WatchManager) Execute(watcherNames []string) error {
	watchers := []*Watcher{}
	for _, name := range watcherNames {
		if watcher, ok := manager.Watchers[name]; ok {
			watchers = append(watchers, watcher)
		} else {
			return errors.New(fmt.Sprintf("invalid watcher '%s'", name))
		}
	}

	for _, watcher := range watchers {
		watcher.Execute()
	}
	return nil
}

// Run all watchers against their listeners.
func (manager *WatchManager) Start() {
	chans := make(map[string][]chan string)
	addChan := func(key, watcher string) chan string {
		if _, ok := chans[key]; !ok {
			chans[key] = []chan string{}
		}
		events := make(chan string)
		chans[key] = append(chans[key], events)
		return events
	}

	for _, watcher := range manager.Watchers {
		for _, key := range watcher.watch {
			if _, ok := manager.listeners[key]; ok {
				events := addChan(key, watcher.Name())
				go watcher.Run(events)
			}
		}
	}

	for key, keyChans := range chans {
		if listener, ok := manager.listeners[key]; ok {
			listener.Start(keyChans)
		}
	}
}

// Stop all watchers.
func (manager *WatchManager) Stop() {
	for _, listener := range manager.listeners {
		listener.Stop()
	}
}
