package main

import (
	"encoding/json"
	"errors"
	"io"
	"net/url"
	"strings"
)

// AddrHost takes an addr string of the form host:port and returns the host
// part.
func AddrHost(addr string) string {
	parts := strings.SplitN(addr, ":", 2)
	return parts[0]
}

// AddrPort takes an addr string of the form host:port and returns the port
// part. If no port exists an empty string is returned.
func AddrPort(addr string) string {
	parts := strings.SplitN(addr, ":", 2)
	if len(parts) > 1 {
		return parts[1]
	}
	return ""
}

// URLScheme parses the URL and returns the scheme. An empty string is returned
// if the URL is not parseable.
func URLScheme(u string) string {
	if parsed, err := url.Parse(u); err == nil {
		return parsed.Scheme
	}
	return ""
}

// URLUsername parses the URL and returns the username. An empty string is
// returned if the URL is not parseable.
func URLUsername(u string) string {
	if parsed, err := url.Parse(u); err == nil && parsed.User != nil {
		return parsed.User.Username()
	}
	return ""
}

// URLPassword parses the URL and returns the password. An empty string is
// returned if the URL is not parseable.
func URLPassword(u string) string {
	if parsed, err := url.Parse(u); err == nil && parsed.User != nil {
		if pass, ok := parsed.User.Password(); ok {
			return pass
		}
	}
	return ""
}

// URLHost parses the URL and returns the host. An empty string is returned
// if the URL is not parseable.
func URLHost(u string) string {
	if parsed, err := url.Parse(u); err == nil {
		return parsed.Host
	}
	return ""
}

// URLPath parses the URL and returns the path. An empty string is returned
// if the URL is not parseable.
func URLPath(u string) string {
	if parsed, err := url.Parse(u); err == nil {
		return parsed.Path
	}
	return ""
}

// URLRawQuery parses the URL and returns the full query string. An empty
// string is returned if the URL is not parseable.
func URLRawQuery(u string) string {
	if parsed, err := url.Parse(u); err == nil {
		return parsed.RawQuery
	}
	return ""
}

// URLQuery parses the URL and query string and returns the first value of the
// named query variable. An empty string is returned if parsing fails or the
// query value does not exist.
func URLQuery(u, name string) string {
	if parsed, err := url.Parse(u); err == nil {
		if query, err := url.ParseQuery(parsed.RawQuery); err == nil {
			return query.Get(name)
		}
	}
	return ""
}

// URLFragment parses the URL and returns the fragment. An empty string is
// returned if the URL is not parseable.
func URLFragment(u string) string {
	if parsed, err := url.Parse(u); err == nil {
		return parsed.Fragment
	}
	return ""
}

// JSON parses the JSON value `item` and returns it.
func JSON(item interface{}) (interface{}, error) {
	var err error
	var data interface{}
	switch item.(type) {
	case []byte:
		err = json.Unmarshal(item.([]byte), &data)
	case string:
		err = json.Unmarshal([]byte(item.(string)), &data)
	case io.Reader:
		dec := json.NewDecoder(item.(io.Reader))
		err = dec.Decode(&data)
	default:
		err = errors.New("item must be string or byte array")
	}
	return data, err
}
