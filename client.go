package main

import (
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

// Join multiple key paths into one. The resulting path will be absolute.
func joinPaths(paths ...string) string {
	path := ""
	for _, part := range paths {
		part = strings.Trim(part, "/")
		if part != "" {
			path = path + "/" + part
		}
	}
	return path
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

// Convert the node tree to a map rooted at the prefix.
func getNodeMap(prefix string, node *etcd.Node) map[string]interface{} {
	prefix = strings.Trim(prefix, "/") + "/"
	path := strings.TrimPrefix(strings.Trim(node.Key, "/"), prefix)
	path = strings.Replace(path, "-", "_", -1)
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

// etcd client wrapper.
type Client struct {
	client *etcd.Client
	prefix string
}

// Create a new client.
func NewClient(config *settings.Settings) (*Client, error) {
	uris := config.StringArrayDflt("uris", []string{})
	if len(uris) == 0 {
		uris = DefaultEtcdURIs
	}
	for n, uri := range uris {
		uris[n] = strings.TrimRight(uri, "/")
	}

	prefix := config.StringDflt("prefix", "")
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

	return &Client{
		client: etcdClient,
		prefix: prefix,
	}, nil
}

// Get a single key and convert it to a map. Returns an empty map if the is not found. Returns an error on failure.
func (c *Client) getOne(prefix, key string) (map[string]interface{}, error) {
	key = joinPaths(prefix, key)
	if response, err := c.client.Get(key, false, true); err == nil {
		return getNodeMap(prefix, response.Node), nil
	} else if etcdErr, ok := err.(*etcd.EtcdError); ok && etcdErr.ErrorCode == 100 {
		return make(map[string]interface{}), nil
	} else {
		return nil, err
	}
}

// Return a series of keys merged into a single value.
func (c *Client) Get(prefix string, keys []string) (map[string]interface{}, error) {
	var err error
	mapping := make(map[string]interface{})
	for _, key := range keys {
		var keyMapping map[string]interface{}
		if keyMapping, err = c.getOne(prefix, key); err == nil {
			mapping = mergemap.Merge(mapping, keyMapping)
		} else {
			break
		}
	}
	return mapping, err
}

// Watch a single prefix for changes.
func (c *Client) watchOne(prefix string, changes chan string, stop chan bool) {
	defer close(changes)
	prefix = strings.Trim(prefix, "/")
	var waitIndex uint64 = 0
	var retryTime int64 = retrySeed

Loop:
	for {
		var err error
		var response *etcd.Response
		if response, err = c.client.Watch(prefix, waitIndex, true, nil, stop); err == nil {
			waitIndex = response.EtcdIndex
			retryTime = retrySeed
			changes <- prefix
		} else if err == etcd.ErrWatchStoppedByUser {
			err = nil
			break
		} else {
			logger.Infof("watch on %s failed, retrying in %.1f seconds", prefix, retryTime/1000)
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
func (c *Client) Watch(prefixes []string, changes chan string, stop chan bool) {
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
		go func(sync syncStore) {
			c.watchOne(prefix, changes, sync.stop)
			close(sync.join)
		}(syncs[n])
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
