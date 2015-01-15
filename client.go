package main

import (
	"github.com/coreos/go-etcd/etcd"
	"github.com/peterbourgon/mergemap"
	"strings"
)

// Make a prefix from a path. The resulting prefix will begin with a / and not end in a /.
func makePrefix(path string) string {
	return "/" + strings.Trim(path, "/")
}

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
func keyName(path string) string {
	parts := strings.Split(path, "/")
	unclean := parts[len(parts)-1]
	return strings.Replace(unclean, "-", "_", -1)
}

// Return a value containing node contents.
func nodeValue(node *etcd.Node) interface{} {
	if node.Dir {
		mapping := make(map[string]interface{})
		for _, child := range node.Nodes {
			name := keyName(child.Key)
			mapping[name] = nodeValue(child)
		}
		return mapping
	} else {
		return node.Value
	}
}

// etcd client wrapper.
type Client struct {
	client *etcd.Client
	prefix string
}

// Internal client creation.
func newClient(etcdClient *etcd.Client, prefix string) *Client {
	return &Client{
		etcdClient,
		makePrefix(prefix),
	}
}

// Create a new client.
func NewClient(uris []string, prefix string) *Client {
	return newClient(etcd.NewClient(uris), prefix)
}

// Create a new client with TLS enabled.
func NewTLSClient(uris []string, prefix string, tlsCert string, tlsKey string, tlsCaCert string) (client *Client, err error) {
	var etcdClient *etcd.Client
	if etcdClient, err = etcd.NewTLSClient(uris, tlsCert, tlsKey, tlsCaCert); err == nil {
		client = newClient(etcdClient, prefix)
	}
	return
}

// Create a mapping rooted at the prefix.
func (c *Client) nodeMapping(prefix string, node *etcd.Node) map[string]interface{} {
	prefix = strings.Trim(prefix, "/") + "/"
	path := strings.TrimPrefix(strings.Trim(node.Key, "/"), prefix)
	path = strings.Replace(path, "-", "_", -1)
	parts := strings.Split(path, "/")
	base := parts[len(parts)-1]
	dir := parts[:len(parts)-1]
	mapping := map[string]interface{}{
		base: nodeValue(node),
	}

	for i := len(dir) - 1; i >= 0; i-- {
		mapping = map[string]interface{}{
			parts[i]: mapping,
		}
	}
	return mapping
}

// Return a key as a map value.
func (c *Client) GetMap(prefix, key string, recursive bool) (map[string]interface{}, error) {
	key = joinPaths(prefix, key)
	if response, err := c.client.Get(key, false, recursive); err == nil {
		logger.Debugf("get key '%s': %v", key, response.Node)
		return c.nodeMapping(prefix, response.Node), nil
	} else {
		logger.Debugf("get key '%s': %s", key, err)
		return nil, err
	}
}

// Return a series of keys merged into a single value.
func (c *Client) GetMaps(prefix string, keys []string, recursive bool) (mapping map[string]interface{}, err error) {
	mapping = make(map[string]interface{})
	for _, key := range keys {
		if nodeMapping, err := c.GetMap(prefix, key, recursive); err == nil {
			mapping = mergemap.Merge(mapping, nodeMapping)
		} else {
			break
		}
	}
	return
}
