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
	path := ""
	for _, part := range paths {
		part = strings.Trim(part, "/")
		if part != "" {
			path = path + "/" + part
		}
	}
	return strings.Trim(path, "/")
}

// Join the `prefix` to the list of `keys` and return the result.
func ResolvePaths(prefix string, keys []string) []string {
	for n, key := range keys {
		keys[n] = JoinPath(prefix, key)
	}
	return keys
}
