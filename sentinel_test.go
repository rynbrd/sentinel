package main

import (
	"errors"
	"gopkg.in/BlueDragonX/go-log.v1"
	"reflect"
	"testing"
	"time"
)

func init() {
	logger.SetLevel(log.LevelDebug)
}

func TestSentinelAdd(t *testing.T) {
	client := &MockClient{}
	ex1 := &MockExecutor{name: "mock1"}
	ex2 := &MockExecutor{name: "mock2"}
	keys1 := []string{"1", "2"}
	keys2 := []string{"2"}
	s := Sentinel{Client: client}

	s.Add(keys1, ex1)
	wantByName := map[string]Executor{"mock1": ex1}
	wantByKey := map[string][]Executor{
		"1": {ex1},
		"2": {ex1},
	}
	if !reflect.DeepEqual(wantByName, s.executorsByName) {
		t.Errorf("%v != %v", wantByName, s.executorsByName)
	}
	if !reflect.DeepEqual(wantByKey, s.executorsByKey) {
		t.Errorf("%v != %v", wantByKey, s.executorsByKey)
	}

	s.Add(keys2, ex2)
	wantByName = map[string]Executor{"mock1": ex1, "mock2": ex2}
	wantByKey = map[string][]Executor{
		"1": {ex1},
		"2": {ex1, ex2},
	}
	if !reflect.DeepEqual(wantByName, s.executorsByName) {
		t.Errorf("%v != %v", wantByName, s.executorsByName)
	}
	if !reflect.DeepEqual(wantByKey, s.executorsByKey) {
		t.Errorf("%v != %v", wantByKey, s.executorsByKey)
	}
}

func TestSentinelExecute(t *testing.T) {
	client := &MockClient{}
	ex1 := &MockExecutor{name: "mock1"}
	ex2 := &MockExecutor{name: "mock2"}
	keys1 := []string{"1", "2"}
	keys2 := []string{"2"}
	s := Sentinel{Client: client}
	s.Add(keys1, ex1)
	s.Add(keys2, ex2)

	if !s.Execute([]string{"mock1"}) {
		t.Error("an executor failed")
	}
	if ex1.Calls != 1 {
		t.Error("executor not called")
	}
	if ex2.Calls != 0 {
		t.Error("executor was called")
	}

	if !s.Execute([]string{"mock1", "mock2"}) {
		t.Error("an executor failed")
	}
	if ex1.Calls != 2 {
		t.Error("executor not called")
	}
	if ex2.Calls != 1 {
		t.Error("executor not called")
	}

	ex1.Error = errors.New("oops!")
	if s.Execute([]string{"mock1", "mock2"}) {
		t.Error("execute succeeded")
	}
	if ex1.Calls != 3 {
		t.Error("executor not called")
	}
	if ex2.Calls != 2 {
		t.Error("executor not called")
	}

	ex1.Error = nil
	if !s.Execute([]string{}) {
		t.Error("an executor failed")
	}
	if ex1.Calls != 4 {
		t.Error("executor not called")
	}
	if ex2.Calls != 3 {
		t.Error("executor not called")
	}

	if s.Execute([]string{"sirnotappearinginthisfilm"}) {
		t.Error("execute succeeded")
	}
	if ex1.Calls != 4 {
		t.Error("executor called")
	}
	if ex2.Calls != 3 {
		t.Error("executor called")
	}
}

func TestSentinelExecuteKey(t *testing.T) {
	client := &MockClient{}
	wantErr := errors.New("error!")
	ex1 := &MockExecutor{name: "mock1"}
	ex2 := &MockExecutor{name: "mock2", Error: wantErr}
	keys1 := []string{"1", "2"}
	keys2 := []string{"2"}
	s := Sentinel{Client: client}
	s.Add(keys1, ex1)
	s.Add(keys2, ex2)

	s.executeKey("1")
	if ex1.Calls != 1 {
		t.Error("executor1 not called")
	}
	if ex2.Calls != 0 {
		t.Error("executor2 called")
	}

	s.executeKey("2")
	if ex1.Calls != 2 {
		t.Error("executor1 not called")
	}
	if ex2.Calls != 1 {
		t.Error("executor2 not called")
	}
}

func TestSentinelRun(t *testing.T) {
	context := map[string]interface{}{
		"sentinel": map[string]interface{}{
			"a": "aye",
			"b": "bee",
		},
	}
	client := &MockClient{GetValue: context}
	ex := &MockExecutor{name: "mock"}
	keys := []string{"sentinel"}
	s := Sentinel{Client: client}
	s.Add(keys, ex)
	stop := make(chan bool)
	join := make(chan struct{})

	go func() {
		s.Run(stop)
		close(join)
	}()
	time.Sleep(1 * time.Millisecond)

	// change causes execution
	client.Changes <- "sentinel"
	time.Sleep(1 * time.Millisecond)
	if ex.Calls != 1 {
		t.Error("executor not called")
	}

	// change to other key causes no execution
	client.Changes <- "beacon"
	time.Sleep(1 * time.Millisecond)
	if ex.Calls != 1 {
		t.Error("executor called")
	}

	stop <- true
	<-join
}
