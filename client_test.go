package main

import (
	"github.com/coreos/go-etcd/etcd"
	"gopkg.in/BlueDragonX/go-settings.v0"
	"io/ioutil"
	"net/http"
	"reflect"
	"strings"
	"testing"
)

const (
	validURI = "http://127.0.0.1:4001"
	invalidURI = "http://127.0.0.1:9999"
)

func getClient(uri string) *Client {
	config := settings.Settings{}
	if uri != "" {
		config.Set("uris", []string{uri})
	}
	config.Set("prefix", "test")
	client, _ := NewClient(&config)
	return client
}

func clientSetUp(t *testing.T, rawClient *etcd.Client) map[string]interface{} {
	var err error
	if _, err = rawClient.SetDir("test", 0); err != nil {
		t.Fatal(err)
	}
	if _, err = rawClient.Set("test/index", "1", 0); err != nil {
		t.Fatal(err)
	}
	if _, err = rawClient.SetDir("test/values", 0); err != nil {
		t.Fatal(err)
	}
	if _, err = rawClient.Set("test/values/a", "aye", 0); err != nil {
		t.Fatal(err)
	}
	if _, err = rawClient.Set("test/values/b", "bee", 0); err != nil {
		t.Fatal(err)
	}

	want := make(map[string]interface{})
	wantValues := make(map[string]interface{})
	wantValues["a"] = "aye"
	wantValues["b"] = "bee"
	want["values"] = wantValues
	want["index"] = "1"
	return want
}

func clientTearDown(t *testing.T, rawClient *etcd.Client) {
	if _, err := rawClient.Delete("test", true); err != nil {
		t.Fatal(err)
	}
}

// Ensure the etcd server does something when we knock.
func TestClientRaw(t *testing.T) {
	if resp, err := http.Get(validURI + "/v2/keys/"); err == nil {
		defer resp.Body.Close()
		if body, err := ioutil.ReadAll(resp.Body); err == nil {
			t.Log(string(body))
		} else {
			t.Error(err)
		}
	} else {
		t.Error(err)
	}

	if _, err := http.Get(invalidURI + "/v2/keys/"); err == nil {
		t.Errorf("no error raised on invalid URI")
	} else {
		t.Logf("%v: %v", reflect.TypeOf(err), err)
	}
}

// Ensure Get returns the correct values.
func TestClientGet(t *testing.T) {
	client := getClient(validURI)
	rawClient := client.client

	// set some keys to play with
	data := clientSetUp(t, rawClient)
	defer clientTearDown(t, rawClient)

	// get `values`
	prefix := "test"
	keys := []string{"values"}
	want := map[string]interface{}{"values": data["values"]}
	if have, err := client.Get(prefix, keys); err == nil {
		if !reflect.DeepEqual(want, have) {
			t.Errorf("keys %v are invalid: %v != %v", keys, want, have)
		}
	} else {
		t.Errorf("keys %v not retrieved: %s\n", keys, err)
	}

	// get `values` and `index`
	keys = []string{"values", "index"}
	want = map[string]interface{}{"values": data["values"], "index": data["index"]}
	if have, err := client.Get(prefix, keys); err == nil {
		if !reflect.DeepEqual(want, have) {
			t.Errorf("keys %v are invalid: %v != %v", keys, want, have)
		}
	} else {
		t.Errorf("keys %v not retrieved: %s\n", keys, err)
	}

	// get `values`, `index`, and `missing`
	keys = []string{"values", "index", "missing"}
	want = map[string]interface{}{"values": data["values"], "index": data["index"]}
	if have, err := client.Get(prefix, keys); err == nil {
		if !reflect.DeepEqual(want, have) {
			t.Errorf("keys %v are invalid: %v != %v", keys, want, have)
		}
	} else {
		t.Errorf("keys %v not retrieved: %s\n", keys, err)
	}

	// get `values` with leading prefix slash
	prefix = "/test"
	keys = []string{"values"}
	want = map[string]interface{}{"values": data["values"]}
	if have, err := client.Get(prefix, keys); err == nil {
		if !reflect.DeepEqual(want, have) {
			t.Errorf("keys %v are invalid: %v != %v", keys, want, have)
		}
	} else {
		t.Errorf("keys %v not retrieved: %s\n", keys, err)
	}

	// get `values` with leading prefix and key slashes
	prefix = "/test"
	keys = []string{"/values"}
	want = map[string]interface{}{"values": data["values"]}
	if have, err := client.Get(prefix, keys); err == nil {
		if !reflect.DeepEqual(want, have) {
			t.Errorf("keys %v are invalid: %v != %v", keys, want, have)
		}
	} else {
		t.Errorf("keys %v not retrieved: %s\n", keys, err)
	}

	// get `values` with leading key slashes
	prefix = "test"
	keys = []string{"/values"}
	want = map[string]interface{}{"values": data["values"]}
	if have, err := client.Get(prefix, keys); err == nil {
		if !reflect.DeepEqual(want, have) {
			t.Errorf("keys %v are invalid: %v != %v", keys, want, have)
		}
	} else {
		t.Errorf("keys %v not retrieved: %s\n", keys, err)
	}

	// get `values` with leading and trailing prefix slash
	prefix = "/test/"
	keys = []string{"values"}
	want = map[string]interface{}{"values": data["values"]}
	if have, err := client.Get(prefix, keys); err == nil {
		if !reflect.DeepEqual(want, have) {
			t.Errorf("keys %v are invalid: %v != %v", keys, want, have)
		}
	} else {
		t.Errorf("keys %v not retrieved: %s\n", keys, err)
	}

	// get `values` with leading and trailing prefix and key slashes
	prefix = "/test/"
	keys = []string{"/values/"}
	want = map[string]interface{}{"values": data["values"]}
	if have, err := client.Get(prefix, keys); err == nil {
		if !reflect.DeepEqual(want, have) {
			t.Errorf("keys %v are invalid: %v != %v", keys, want, have)
		}
	} else {
		t.Errorf("keys %v not retrieved: %s\n", keys, err)
	}

	// get `values` with leading key slashes
	prefix = "test/"
	keys = []string{"/values/"}
	want = map[string]interface{}{"values": data["values"]}
	if have, err := client.Get(prefix, keys); err == nil {
		if !reflect.DeepEqual(want, have) {
			t.Errorf("keys %v are invalid: %v != %v", keys, want, have)
		}
	} else {
		t.Errorf("keys %v not retrieved: %s\n", keys, err)
	}
}

// Ensure the client returns properly when get fails.
func TestClientGetFailure(t *testing.T) {
	// server down should return an error
	client := getClient(invalidURI)
	if value, err := client.Get("/test", []string{"/values"}); err == nil {
		t.Errorf("received a value from etcd: %v", value)
	} else {
		t.Logf("received %v: %v", reflect.TypeOf(err), err)
	}

	// missing key should return an empty map
	client = getClient(validURI)
	if value, err := client.Get("/test", []string{"/values"}); err == nil {
		t.Logf("received a value from etcd: %v", value)
	} else {
		t.Errorf("received %v: %v", reflect.TypeOf(err), err)
	}
}

// Ensure we URIs with and without trailing slashes work.
func TestClientTrailingSlash(t *testing.T) {
	client := getClient(validURI)
	clientSetUp(t, client.client)
	defer clientTearDown(t, client.client)

	// test without a trailing slash
	uri := strings.TrimRight(validURI, "/")
	client = getClient(uri)
	if _, err := client.Get("test", []string{"/values"}); err != nil {
		t.Errorf("URI without trailing slash failed: %s", err)
	}

	// test with a trailing slash
	uri = uri + "/"
	client = getClient(uri)
	if _, err := client.Get("test", []string{"/values"}); err != nil {
		t.Errorf("URI with trailing slash failed: %s", err)
	}
}
