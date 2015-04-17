package main

import (
	"fmt"
	"os"
	"strings"
)

// Print the formatted error and exit.
func Fatalf(format string, a ...interface{}) {
	fmt.Fprintf(os.Stderr, format, a...)
	os.Exit(1)
}

// Join multiple key paths into one. The resulting path will be absolute.
func JoinPath(paths ...string) string {
	parts := []string{}
	for _, path := range paths {
		path = CleanPath(path)
		if path != "" {
			parts = append(parts, path)
		}
	}
	return strings.Join(parts, "/")
}

// CleanPath removes leading, trailing, and duplicate slashes.
func CleanPath(path string) string {
	parts := []string{}
	for _, part := range strings.Split(path, "/") {
		if part != "" {
			parts = append(parts, part)
		}
	}
	return strings.Join(parts, "/")
}

// Join the `prefix` to the list of `keys` and return the result.
func ResolvePaths(prefix string, keys []string) []string {
	for n, key := range keys {
		keys[n] = JoinPath(prefix, key)
	}
	return keys
}
