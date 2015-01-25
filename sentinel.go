package main

import (
	"fmt"
)

type Sentinel struct {
	Client          Client
	executorsByName map[string]*Executor
	executorsByKey  map[string][]*Executor
}

// Add an `executor` for the provided `keys`.
func (s *Sentinel) Add(keys []string, executor *Executor) {
	if s.executorsByName == nil {
		s.executorsByName = make(map[string]*Executor)
	}
	if s.executorsByKey == nil {
		s.executorsByKey = make(map[string][]*Executor)
	}
	s.executorsByName[executor.Name] = executor
	for _, key := range keys {
		logger.Debugf("execute %s on '%s'", executor.Name, key)
		executorArray, ok := s.executorsByKey[key]
		if !ok {
			executorArray = make([]*Executor, 0, 1)
		}
		s.executorsByKey[key] = append(executorArray, executor)
	}
}

// Look up an executor by name and execute it.
func (s *Sentinel) ExecByName(name string) error {
	if executor, ok := s.executorsByName[name]; ok {
		return executor.Execute(s.Client)
	} else {
		return fmt.Errorf("executor %s not found", name)
	}
}

// Look up a executors by prefix and execute them.
func (s *Sentinel) ExecByPrefix(prefix string) []error {
	executors, ok := s.executorsByKey[prefix]
	if !ok {
		return []error{}
	}

	errors := make([]error, 0)
	for _, executor := range executors {
		if err := executor.Execute(s.Client); err != nil {
			errors = append(errors, err)
		}
	}
	return errors
}

// Get the prefixes we're configured to watch.
func (s *Sentinel) getPrefixes() []string {
	prefixes := make([]string, 0, len(s.executorsByKey))
	for prefix := range s.executorsByKey {
		prefixes = append(prefixes, prefix)
	}
	return prefixes
}

func (s *Sentinel) Run(stop chan bool) {
	changes := make(chan string, 10)
	watchStop := make(chan bool)
	watchJoin := make(chan struct{})
	go func() {
		s.Client.Watch(s.getPrefixes(), changes, watchStop)
		close(watchJoin)
	}()

Loop:
	for {
		select {
		case <-stop:
			watchStop <- true
			<-watchJoin
			break Loop
		case prefix := <-changes:
			logger.Debugf("prefix '%s' changed", prefix)
			if errs := s.ExecByPrefix(prefix); len(errs) > 0 {
				for _, err := range errs {
					logger.Error(err.Error())
				}
			}
		}
	}
}
