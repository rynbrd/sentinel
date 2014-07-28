package main

import (
	"errors"
	"fmt"
	"gopkg.in/BlueDragonX/simplelog.v1"
	"os/exec"
	"strings"
)

// A watcher renders templates and executes command when key changes are detected.
type Watcher struct {
	name      string
	watch    []string
	context  []string
	renderer  *Renderer
	command   []string
	client    *Client
	logger    *simplelog.Logger
}

// Create a new watcher.
func NewWatcher(name string, watch, context []string, renderer *Renderer, command []string, client *Client, logger *simplelog.Logger) *Watcher {
	return &Watcher{
		name,
		watch,
		context,
		renderer,
		command,
		client,
		logger,
	}
}

// Run the watcher command.
func (watcher *Watcher) runCommand() error {
	if len(watcher.command) == 0 {
		watcher.logger.Debug("%s has no command, skipping", watcher.name)
		return nil
	}

	cmdName := watcher.command[0]
	cmdArgs := watcher.command[1:]
	command := exec.Command(cmdName, cmdArgs...)

	watcher.logger.Debug("%s calling command", watcher.name)
	out, err := command.CombinedOutput()
	if err != nil {
		watcher.logger.Error("%s cmd failed: %s", watcher.name, err)
	}
	outStr := string(out)
	if outStr != "" {
		lines := strings.Split(outStr, "\n")
		for _, line := range lines {
			if err == nil {
				watcher.logger.Debug("%s cmd: %s", watcher.name, line)
			} else {
				watcher.logger.Warn("%s cmd: %s", watcher.name, line)
			}
		}
	}
	return err
}

// Return the name of the watcher.
func (watcher *Watcher) Name() string {
	return watcher.name
}

// Execute the watcher as if an event was receieved.
func (watcher *Watcher) Execute() error {
	context, err := watcher.client.GetMaps(watcher.context, true)
	if err != nil {
		watcher.logger.Error("%s failed to retrieve context: %s", watcher.Name(), err)
		return err
	}

	changed := true
	if watcher.renderer != nil {
		changed, err = watcher.renderer.Render(context)
		if err != nil {
			watcher.logger.Error("%s failed to render: %s", watcher.Name(), err)
			return err
		}
	}

	if changed {
		err = watcher.runCommand()
		if err != nil {
			watcher.logger.Error("%s failed to run command: %s", watcher.Name(), err)
			return err
		}
		watcher.logger.Info("%s executed", watcher.Name())
	} else {
		watcher.logger.Info("%s skipped execution", watcher.Name())
	}
	return nil
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
	watchers map[string]*Watcher
	listeners map[string]*Listener
	client *Client
	logger *simplelog.Logger
}

// Create a new watch manager.
func NewWatchManager(watchers []*Watcher, client *Client, logger *simplelog.Logger) *WatchManager {
	manager := &WatchManager{
		make(map[string]*Watcher),
		make(map[string]*Listener),
		client,
		logger,
	}

	for _, watcher := range watchers {
		manager.watchers[watcher.Name()] = watcher
		for _, key := range watcher.watch {
			if _, have := manager.listeners[key]; !have {
				manager.listeners[key] = NewListener(watcher.client.prefix, key, client, logger)
			}
		}
	}
	return manager
}

// Execute the named watchers.
func (manager *WatchManager) Execute(watcherNames []string) error {
	watchers := []*Watcher{}
	for _, name := range watcherNames {
		if watcher, ok := manager.watchers[name]; ok {
			watchers = append(watchers, watcher)
		} else {
			return errors.New(fmt.Sprintf("invalid watcher '%s'", name))
		}
	}

	for _, watcher := range watchers {
		if err := watcher.Execute(); err != nil {
			return err
		}
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

	for _, watcher := range manager.watchers {
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
