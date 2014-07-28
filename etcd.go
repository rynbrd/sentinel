package main

import (
	"github.com/coreos/go-etcd/etcd"
	"strings"
)

// Return the base name of a node key.
func GetKeyName(key string) string {
	keyParts := strings.Split(key, "/")
	return keyParts[len(keyParts)-1]
}

// Find a node with the provided key name.
func FindNode(name string, node etcd.Node) *etcd.Node {
	for _, node := range node.Nodes {
		if GetKeyName(node.Key) == name {
			return node
		}
	}
	return nil
}
