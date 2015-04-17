package main

import (
	"reflect"
	"testing"
)

func TestJoinPath(t *testing.T) {
	type joinCheck struct {
		want  string
		paths []string
	}

	joinChecks := []joinCheck{
		{"test/value", []string{"test", "value"}},
		{"test/value", []string{"/test", "/value"}},
		{"test/value", []string{"/test/", "/value/"}},
		{"test/value", []string{"test/", "value/"}},
		{"test/value", []string{"/test", "value/"}},
		{"test/value", []string{"test/", "/value"}},
		{"test/value", []string{"/test/", "value"}},
		{"test/value", []string{"test", "/value/"}},
		{"test/value/a", []string{"test", "value", "a"}},
		{"test/value/a", []string{"test/value", "a"}},
		{"test/value/a", []string{"test//value", "a"}},
		{"test/value/a", []string{"test//value/", "a"}},
		{"test/value/a", []string{"/test//value/", "/a"}},
	}

	for _, check := range joinChecks {
		if have := JoinPath(check.paths...); check.want != have {
			t.Errorf("path '%s' != '%s'", check.want, have)
		}
	}
}

func TestCleanPath(t *testing.T) {
	cleanChecks := [][2]string{
		{"test/value/a", "/test/value/a"},
		{"test/value/a", "/test/value/a/"},
		{"test/value/a", "/test/value/a/"},
		{"test/value/a", "test//value/a"},
		{"test/value/a", "test/value//a"},
		{"test/value/a", "//test/////////value//////a//"},
	}

	for _, check := range cleanChecks {
		if have := CleanPath(check[1]); check[0] != have {
			t.Errorf("path '%s' != '%s'", check[0], have)
		}
	}
}

func TestResolvePaths(t *testing.T) {
	want := []string{"test/value/a", "test/value/b"}
	prefix := "test"
	keys := []string{"value/a", "value/b"}
	have := ResolvePaths(prefix, keys)
	if !reflect.DeepEqual(want, have) {
		t.Errorf("paths %v != %v", want, have)
	}
}
