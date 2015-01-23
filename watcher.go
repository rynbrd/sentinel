package main

import (
	"errors"
	"fmt"
	"gopkg.in/BlueDragonX/go-settings.v0"
	"os/exec"
	"strings"
)

// A watcher renders templates and executes command when key changes are detected.
type Watcher struct {
	name     string
	prefix   string
	watch    []string
	context  []string
	renderer *Renderer
	command  []string
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
	renderer := &Renderer{tpls}

	return &Watcher{
		name:     name,
		prefix:   prefix,
		watch:    watch,
		context:  context,
		renderer: renderer,
		command:  command,
		client:   client,
	}, nil
}

// Run the watcher command.
func (watcher *Watcher) runCommand() error {
	if len(watcher.command) == 0 {
		logger.Debugf("%s has no command, skipping", watcher.name)
		return nil
	}

	cmdName := watcher.command[0]
	cmdArgs := watcher.command[1:]
	command := exec.Command(cmdName, cmdArgs...)

	logger.Debugf("%s calling command", watcher.name)
	out, err := command.CombinedOutput()
	if err != nil {
		logger.Errorf("%s cmd failed: %s", watcher.name, err)
	}
	outStr := string(out)
	if outStr != "" {
		lines := strings.Split(outStr, "\n")
		for _, line := range lines {
			if err == nil {
				logger.Debugf("%s cmd: %s", watcher.name, line)
			} else {
				logger.Errorf("%s cmd: %s", watcher.name, line)
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
	context, err := watcher.client.Get(watcher.prefix, watcher.context)
	if err != nil {
		logger.Errorf("%s failed to retrieve context: %s", watcher.Name(), err)
		return err
	}

	logger.Debugf("context: %v\n", context)
	changed := true
	if watcher.renderer != nil {
		changed, err = watcher.renderer.Render(context)
		if err != nil {
			logger.Errorf("%s failed to render: %s", watcher.Name(), err)
			return err
		}
	}

	if changed {
		err = watcher.runCommand()
		if err != nil {
			logger.Errorf("%s failed to run command: %s", watcher.Name(), err)
			return err
		}
		logger.Infof("%s executed", watcher.Name())
	} else {
		logger.Infof("%s skipped execution", watcher.Name())
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
