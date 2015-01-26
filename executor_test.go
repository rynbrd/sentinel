package main

import (
	"io/ioutil"
	"os"
	"path"
	"reflect"
	"testing"
)

type MockExecutor struct {
	name  string
	Calls int
	Error error
}

func (ex *MockExecutor) Name() string {
	return ex.name
}

func (ex *MockExecutor) Execute(client Client) error {
	ex.Calls++
	return ex.Error
}

type ExecutorTestCase struct {
	T         *testing.T
	Client    *MockClient
	Template  Template
	Directory string
	Context   map[string]interface{}
	ContextA  map[string]interface{}
	ContextB  map[string]interface{}
}

func NewExecutorTestCase(t *testing.T) *ExecutorTestCase {
	tc := ExecutorTestCase{}

	var err error
	if tc.Directory, err = ioutil.TempDir("", "sentinel_test_"); err != nil {
		t.Fatalf("failed to create temp directory: %s", err)
	}

	tc.Template = Template{
		Src:  path.Join(tc.Directory, "src"),
		Dest: path.Join(tc.Directory, "dest"),
	}

	tc.ContextA = map[string]interface{}{"value": "a"}
	tc.ContextB = map[string]interface{}{"value": "b"}
	tc.Context = map[string]interface{}{
		"sentinel": map[string]interface{}{
			"context_a": tc.ContextA,
			"context_b": tc.ContextB,
		},
	}
	tc.Client = &MockClient{GetValue: tc.Context}

	return &tc
}

func (tc *ExecutorTestCase) Close() error {
	return os.RemoveAll(tc.Directory)
}

func TestExecutorEmpty(t *testing.T) {
	tc := NewExecutorTestCase(t)
	defer tc.Close()

	exec := TemplateExecutor{
		name:    "test",
		prefix:  "sentinel",
		context: []string{"sentinel/context_a"},
	}

	if err := exec.Execute(tc.Client); err != nil {
		t.Error(err)
	}
}

func TestExecutorCommand(t *testing.T) {
	tc := NewExecutorTestCase(t)
	defer tc.Close()
	out := path.Join(tc.Directory, "out")

	exec := TemplateExecutor{
		name:    "test",
		prefix:  "sentinel",
		context: []string{},
		Command: []string{"bash", "-c", "echo hello > " + out},
	}

	if err := exec.Execute(tc.Client); err != nil {
		t.Errorf("failed to execute: %s", err)
	}

	want := []byte("hello\n")
	if have, err := ioutil.ReadFile(out); err == nil {
		t.Logf("out: %s", string(have))
		if !reflect.DeepEqual(want, have) {
			t.Error("command output incorrect")
		}
	} else {
		t.Errorf("command output incorrect")
	}
}

func TestExecutorSingleContext(t *testing.T) {
	tc := NewExecutorTestCase(t)
	defer tc.Close()

	tc.Client.GetValue = map[string]interface{}{
		"sentinel": map[string]interface{}{
			"context_a": tc.ContextA,
		},
	}

	exec := TemplateExecutor{
		name:      "test",
		prefix:    "sentinel",
		context:   []string{"sentinel/context_a"},
		Templates: []Template{tc.Template},
	}

	var err error
	src := []byte("value: {{ .context_a.value }}\n")
	dest := []byte("value: a\n")
	if err = ioutil.WriteFile(tc.Template.Src, src, 0600); err != nil {
		t.Fatal(err)
	}

	if err := exec.Execute(tc.Client); err != nil {
		t.Errorf("failed to execute: %s", err)
	}
	if have, err := ioutil.ReadFile(tc.Template.Dest); err == nil {
		t.Log(string(have))
		if !reflect.DeepEqual(dest, have) {
			t.Error("template destination incorrectly rendered")
		}
	} else {
		t.Errorf("template destination not rendered: %s", err)
	}
}

func TestExecutorMultiContext(t *testing.T) {
	tc := NewExecutorTestCase(t)
	defer tc.Close()

	exec := TemplateExecutor{
		name:      "test",
		prefix:    "sentinel",
		context:   []string{"sentinel/context_a"},
		Templates: []Template{tc.Template},
	}

	var err error
	src := []byte("value-a: {{ .context_a.value }}\nvalue-b: {{ .context_b.value }}\n")
	dest := []byte("value-a: a\nvalue-b: b\n")
	if err = ioutil.WriteFile(tc.Template.Src, src, 0600); err != nil {
		t.Fatal(err)
	}

	if err := exec.Execute(tc.Client); err != nil {
		t.Errorf("failed to execute: %s", err)
	}
	if have, err := ioutil.ReadFile(tc.Template.Dest); err == nil {
		t.Log(string(have))
		if !reflect.DeepEqual(dest, have) {
			t.Error("template destination incorrectly rendered")
		}
	} else {
		t.Errorf("template destination not rendered: %s", err)
	}
}

func TestExecutorChanged(t *testing.T) {
	tc := NewExecutorTestCase(t)
	defer tc.Close()
	out := path.Join(tc.Directory, "out")

	exec := TemplateExecutor{
		name:      "test",
		prefix:    "sentinel",
		context:   []string{"sentinel/context_a"},
		Templates: []Template{tc.Template},
		Command:   []string{"bash", "-c", "echo hello > " + out},
	}

	var err error
	src := []byte("value-a: {{ .context_a.value }}\nvalue-b: {{ .context_b.value }}\n")
	if err = ioutil.WriteFile(tc.Template.Src, src, 0600); err != nil {
		t.Fatal(err)
	}

	if err := exec.Execute(tc.Client); err != nil {
		t.Errorf("failed to execute: %s", err)
	}

	want := []byte("hello\n")
	if have, err := ioutil.ReadFile(out); err == nil {
		if !reflect.DeepEqual(want, have) {
			t.Error("command output incorrect")
		}
	} else {
		t.Errorf("command output incorrect")
	}

	os.Remove(out)
	if err := exec.Execute(tc.Client); err != nil {
		t.Errorf("failed to execute: %s", err)
	}

	if _, err := os.Stat(out); !os.IsNotExist(err) {
		t.Error("command executed with no template change")
	}
}
