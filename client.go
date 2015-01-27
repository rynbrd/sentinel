package main

import (
	"encoding/json"
	"github.com/coreos/go-etcd/etcd"
	"github.com/peterbourgon/mergemap"
	"gopkg.in/BlueDragonX/go-settings.v0"
	"strings"
	"time"
)

const (
	retrySeed   = 500
	retryMax    = 30000
	retryFactor = 1.5
)

var DefaultEtcdURIs []string = []string{"http://172.17.42.1:4001/"}

// Configuration server client interface.
type Client interface {
	// Wait for the server to become available. The wait can be stopped by
	// sending a value to `stop` or closing it. Return true if the server came
	// online or false if the wait was canceled.
	Wait(stop chan bool) bool

	// Recursively retrieve the values for a group of `keys`. The values are
	// merged into a map of values structured as a tree.
	Get(keys []string) (map[string]interface{}, error)

	// Recursively watch each prefix in `prefixes` for changes. Send the name
	// of the changed prefix to the `changes` channel. Stop watching and exit
	// when `stop` receives `true`. Wait for the server to become available if
	// it isn't. Each failed watch attempt will be followed by an increasingly
	// longer period of sleep.
	Watch(prefixes []string, changes chan string, stop chan bool)
}

// Return the base key name for a key path.
func getKeyName(path string) string {
	parts := strings.Split(path, "/")
	unclean := parts[len(parts)-1]
	return strings.Replace(unclean, "-", "_", -1)
}

// Return a value containing node contents.
func getNodeValue(node *etcd.Node) interface{} {
	if node.Dir {
		mapping := make(map[string]interface{})
		for _, child := range node.Nodes {
			name := getKeyName(child.Key)
			mapping[name] = getNodeValue(child)
		}
		return mapping
	} else {
		return node.Value
	}
}

// Convert the node tree to a map.
func getNodeMap(node *etcd.Node) map[string]interface{} {
	path := strings.Replace(strings.Trim(node.Key, "/"), "-", "_", -1)
	parts := strings.Split(path, "/")
	base := parts[len(parts)-1]
	dir := parts[:len(parts)-1]
	mapping := map[string]interface{}{
		base: getNodeValue(node),
	}

	for i := len(dir) - 1; i >= 0; i-- {
		mapping = map[string]interface{}{
			parts[i]: mapping,
		}
	}
	return mapping
}

// An etcd client implementation.
type EtcdClient struct {
	client *etcd.Client
}

// Create a new etcd client.
func NewEtcdClient(config *settings.Settings) (*EtcdClient, error) {
	uris := config.StringArrayDflt("uris", []string{})
	if len(uris) == 0 {
		uris = DefaultEtcdURIs
	}
	for n, uri := range uris {
		uris[n] = strings.TrimRight(uri, "/")
	}

	tlsKey := config.StringDflt("tls-key", "")
	tlsCert := config.StringDflt("tls-cert", "")
	tlsCaCert := config.StringDflt("tls-ca-cert", "")

	var err error
	var etcdClient *etcd.Client
	if tlsKey != "" && tlsCert != "" && tlsCaCert != "" {
		if etcdClient, err = etcd.NewTLSClient(uris, tlsCert, tlsKey, tlsCaCert); err != nil {
			return nil, err
		}
	} else {
		etcdClient = etcd.NewClient(uris)
	}

	return &EtcdClient{
		client: etcdClient,
	}, nil
}

// Wait for the server to become available. The wait can be stopped by sending
// a value to `stop` or closing it. Return true if the server came online or
// false if the wait was canceled.
func (c *EtcdClient) Wait(stop chan bool) bool {
	var retryTime int64 = retrySeed
	for {
		if _, err := c.client.Get("/", false, false); err == nil {
			logger.Debug("connected to etcd")
			break
		} else {
			logger.Infof("waiting %.1f seconds for etcd", float64(retryTime)/1000.0)
			logger.Debugf("error was: %s", err)

			select {
			case <-time.After(time.Duration(retryTime) * time.Millisecond):
			case <-stop:
				return false
			}

			if retryTime < retryMax {
				retryTime = int64(float64(retryTime) * float64(retryFactor))
			} else {
				retryTime = retryMax
			}
		}
	}
	return true
}

// Get a single key and convert it to a map. Returns an empty map if the is not
// found. Returns an error on failure.
func (c *EtcdClient) getOne(key string) (map[string]interface{}, error) {
	if response, err := c.client.Get(key, false, true); err == nil {
		item := getNodeMap(response.Node)
		return item, nil
	} else if etcdErr, ok := err.(*etcd.EtcdError); ok && etcdErr.ErrorCode == 100 {
		return make(map[string]interface{}), nil
	} else {
		return nil, err
	}
}

// Get a group of keys rooted and merge them into a single map.
func (c *EtcdClient) Get(keys []string) (map[string]interface{}, error) {
	var err error
	mapping := make(map[string]interface{})
	for _, key := range keys {
		var keyMapping map[string]interface{}
		if keyMapping, err = c.getOne(key); err == nil {
			mapping = mergemap.Merge(mapping, keyMapping)
		} else {
			break
		}
	}
	return mapping, err
}

// Watch a single prefix for changes.
func (c *EtcdClient) watchOne(prefix string, changes chan string, stop chan bool) {
	prefix = strings.Trim(prefix, "/")
	var waitIndex uint64 = 0
	var retryTime int64 = retrySeed
	logger.Debugf("watching %s for changes", prefix)

Loop:
	for {
		var err error
		var response *etcd.Response
		if response, err = c.client.Watch(prefix, waitIndex, true, nil, stop); err == nil {
			waitIndex = response.EtcdIndex + 1
			retryTime = retrySeed
			logger.Debugf("prefix %s changed, index was %d, action was %s", prefix, response.EtcdIndex, response.Action)
			changes <- prefix
		} else if _, ok := err.(*json.SyntaxError); ok {
			// This is caused by the connection timing out thus cutting the
			// stream the JSON encoder is readong from. I would expect this to
			// be common on HTTP long-pulls especially if a proxy is involved.
			// On the other hand I am unsure whether changes to the go-etcd
			// client library or the etcd server itself would mitigate this
			// issue when directly connected. Either way we will retry the
			// watch using the same index so as not to miss any changes.
			logger.Debugf("watch on %s timed out, retrying immediately")
		} else if err == etcd.ErrWatchStoppedByUser {
			err = nil
			break
		} else if etcdErr, ok := err.(*etcd.EtcdError); ok && etcdErr.ErrorCode == 401 {
			// This will happen if we lose connectivity for long enough that
			// etcd starts clearing the history. This should happen if we miss
			// 1000 events.
			logger.Errorf("watch on %s index %s cleared, resetting to 0")
			waitIndex = 0
		} else {
			logger.Errorf("watch on %s failed, retrying in %.1f seconds", prefix, float64(retryTime)/1000)
			logger.Debugf("error was: %s", err)

			select {
			case <-time.After(time.Duration(retryTime) * time.Millisecond):
			case <-stop:
				break Loop
			}

			if retryTime < retryMax {
				retryTime = int64(float64(retryTime) * float64(retryFactor))
			} else {
				retryTime = retryMax
			}
		}
	}
}

// Recursively watch each prefix in `prefixes` for changes. Send the name of
// changed prefix to the `changes` channel. Stop watching and exit when `stop`
// receives `true`. Wait for the server to become available if it isn't. Each
// failed attempt will be followed by an increasingly longer period of sleep.
func (c *EtcdClient) Watch(prefixes []string, changes chan string, stop chan bool) {
	defer close(changes)
	type syncStore struct {
		stop chan bool
		join chan bool
	}

	syncs := make([]syncStore, len(prefixes))
	for n, prefix := range prefixes {
		syncs[n] = syncStore{
			make(chan bool),
			make(chan bool),
		}
		go func(prefix string, sync syncStore) {
			c.watchOne(prefix, changes, sync.stop)
			close(sync.join)
		}(prefix, syncs[n])
	}

	<-stop
	for _, sync := range syncs {
		sync.stop <- true
	}
	for _, sync := range syncs {
		select {
		case <-sync.join:
		case <-time.After(200 * time.Millisecond):
		}
	}
}
