package main

import (
	"github.com/BlueDragonX/go-etcd/etcd"
	"gopkg.in/BlueDragonX/go-settings.v0"
	"io/ioutil"
	"net/http"
	"reflect"
	"strings"
	"testing"
	"time"
)

const (
	validURI   = "http://127.0.0.1:4001"
	invalidURI = "http://127.0.0.1:9999"
)

type MockClient struct {
	WaitFor  time.Duration
	GetValue map[string]interface{}
	GetError error
	Changes  chan string
}

func (mc *MockClient) Wait(stop chan bool) bool {
	select {
	case <-time.After(mc.WaitFor):
	case <-stop:
		return false
	}
	return true
}

func (mc *MockClient) Get(keys []string) (map[string]interface{}, error) {
	return mc.GetValue, mc.GetError
}

func (mc *MockClient) Watch(prefixes []string, changes chan string, stop chan bool) {
	mc.Changes = changes
	<-stop
}
func getEtcdClient(t *testing.T, uri string) *EtcdClient {
	config := settings.Settings{}
	if uri != "" {
		config.Set("uris", []string{uri})
	}
	config.Set("prefix", "test")
	client, err := NewEtcdClient(&config)
	if err != nil {
		t.Fatal(err)
	}
	return client
}

func etcdClientSetUp(t *testing.T, rawClient *etcd.Client) map[string]interface{} {
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

func etcdClientTearDown(t *testing.T, rawClient *etcd.Client) {
	if _, err := rawClient.Delete("test", true); err != nil {
		t.Fatal(err)
	}
}

// Ensure the etcd server does something when we knock.
func TestEtcdClientRaw(t *testing.T) {
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
func TestEtcdClientGet(t *testing.T) {
	client := getEtcdClient(t, validURI)
	rawClient := client.client

	// set some keys to play with
	data := etcdClientSetUp(t, rawClient)
	defer etcdClientTearDown(t, rawClient)

	type getCheck struct {
		keys []string
		want interface{}
	}

	getChecks := []getCheck{
		{
			[]string{"test/values"},
			map[string]interface{}{
				"test": map[string]interface{}{
					"values": data["values"],
				},
			},
		},
		{
			[]string{"test/values", "test/index"},
			map[string]interface{}{
				"test": map[string]interface{}{
					"values": data["values"],
					"index":  data["index"],
				},
			},
		},
		{
			[]string{"test/values", "test/index", "test/missing"},
			map[string]interface{}{
				"test": map[string]interface{}{
					"values": data["values"],
					"index":  data["index"],
				},
			},
		},
		{
			[]string{"/test/values"},
			map[string]interface{}{
				"test": map[string]interface{}{
					"values": data["values"],
				},
			},
		},
		{
			[]string{"/test/values/"},
			map[string]interface{}{
				"test": map[string]interface{}{
					"values": data["values"],
				},
			},
		},
	}

	for _, check := range getChecks {
		if have, err := client.Get(check.keys); err == nil {
			if !reflect.DeepEqual(check.want, have) {
				t.Errorf("keys %v are invalid: %v != %v", check.keys, check.want, have)
			}
		} else {
			t.Errorf("keys %v not retrieved: %s\n", check.keys, err)
		}
	}
}

// Ensure the client returns properly when get fails.
func TestEtcdClientGetFailure(t *testing.T) {
	// server down should return an error
	client := getEtcdClient(t, invalidURI)
	if value, err := client.Get([]string{"test/values"}); err == nil {
		t.Errorf("received a value from etcd: %v", value)
	} else {
		t.Logf("received %v: %v", reflect.TypeOf(err), err)
	}

	// missing key should return an empty map
	client = getEtcdClient(t, validURI)
	if value, err := client.Get([]string{"test/values"}); err == nil {
		t.Logf("received a value from etcd: %v", value)
	} else {
		t.Errorf("received %v: %v", reflect.TypeOf(err), err)
	}
}

// Ensure the client watches properly.
func TestEtcdClientWatch(t *testing.T) {
	client := getEtcdClient(t, validURI)
	rawClient := client.client
	etcdClientSetUp(t, rawClient)
	defer etcdClientTearDown(t, rawClient)

	// server down should return an error
	join := make(chan bool)
	changes := make(chan string)
	stop := make(chan bool)
	go func() {
		client.Watch([]string{"test/index"}, changes, stop)
		close(join)
	}()

	time.Sleep(50 * time.Millisecond)
	go func() {
		rawClient.Set("/test/index", "2", 0)
	}()

	if key := <-changes; key != "test/index" {
		t.Errorf("changed key is '%s' not 'test/index'", key)
	} else {
		t.Log("changed key is '%s'\n", key)
	}
	stop <- true
	<-join
}

// Ensure we URIs with and without trailing slashes work.
func TestEtcdClientTrailingSlash(t *testing.T) {
	client := getEtcdClient(t, validURI)
	etcdClientSetUp(t, client.client)
	defer etcdClientTearDown(t, client.client)

	// test without a trailing slash
	uri := strings.TrimRight(validURI, "/")
	client = getEtcdClient(t, uri)
	if _, err := client.Get([]string{"test/values"}); err != nil {
		t.Errorf("URI without trailing slash failed: %s", err)
	}

	// test with a trailing slash
	uri = uri + "/"
	client = getEtcdClient(t, uri)
	if _, err := client.Get([]string{"test/values"}); err != nil {
		t.Errorf("URI with trailing slash failed: %s", err)
	}
}
