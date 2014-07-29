package main

import (
	"errors"
	"gopkg.in/BlueDragonX/simplelog.v1"
	"gopkg.in/BlueDragonX/yamlcfg.v1"
)

const (
	DefaultEtcdURI       = "http://172.17.42.1:4001/"
	DefaultEtcdPrefix    = ""
	DefaultEtcdTLSKey    = ""
	DefaultEtcdTLSCert   = ""
	DefaultEtcdTLSCACert = ""
)

// Store etcd related configuration.
type EtcdConfig struct {
	URIs      []string
	Prefix    string
	TLSKey    string
	TLSCert   string
	TLSCACert string
}

//Get default etcd config.
func DefaultEtcdConfig() EtcdConfig {
	return EtcdConfig{
		[]string{DefaultEtcdURI},
		DefaultEtcdPrefix,
		DefaultEtcdTLSKey,
		DefaultEtcdTLSCert,
		DefaultEtcdTLSCACert,
	}
}

// Create a client from the config.
func (cfg *EtcdConfig) CreateClient(logger *simplelog.Logger) (client *Client, err error) {
	if cfg.IsTLS() {
		client, err = NewTLSClient(cfg.URIs, cfg.Prefix, logger, cfg.TLSCert, cfg.TLSKey, cfg.TLSCACert)
	} else {
		client = NewClient(cfg.URIs, cfg.Prefix, logger)
	}
	return
}

// SetYAML parses the YAML tree into the object.
func (cfg *EtcdConfig) SetYAML(tag string, data interface{}) bool {
	yamlcfg.AssertIsMap("etcd", data)
	cfg.URIs = yamlcfg.GetStringArray(data, "uris", []string{})

	uri := yamlcfg.GetString(data, "uri", "")
	if uri != "" {
		cfg.URIs = append(cfg.URIs, uri)
	}
	if len(cfg.URIs) == 0 {
		cfg.URIs = append(cfg.URIs, DefaultEtcdURI)
	}

	cfg.Prefix = yamlcfg.GetString(data, "prefix", DefaultEtcdPrefix)
	cfg.TLSKey = yamlcfg.GetString(data, "tls-key", DefaultEtcdTLSKey)
	cfg.TLSCert = yamlcfg.GetString(data, "tls-cert", DefaultEtcdTLSCert)
	cfg.TLSCACert = yamlcfg.GetString(data, "tls-ca-cert", DefaultEtcdTLSCACert)
	return true
}

// Validate the configuration.
func (cfg *EtcdConfig) Validate() []error {
	errs := []error{}
	if len(cfg.URIs) == 0 || cfg.URIs[0] == "" {
		errs = append(errs, errors.New("invalid value for etcd.uris"))
	}
	if cfg.Prefix == "" {
		errs = append(errs, errors.New("invalid value for etcd.prefix"))
	}
	if cfg.IsTLS() {
		if !fileIsReadable(cfg.TLSKey) {
			errs = append(errs, errors.New("invalid etcd.tls-key: file is not readable"))
		}
		if !fileIsReadable(cfg.TLSCert) {
			errs = append(errs, errors.New("invalid etcd.tls-cert: file is not readable"))
		}
		if !fileIsReadable(cfg.TLSCACert) {
			errs = append(errs, errors.New("invalid etcd.tls-ca-cert: file is not readable"))
		}
	}
	return errs
}

// Check if TLS is enabled.
func (cfg EtcdConfig) IsTLS() bool {
	return cfg.TLSKey != "" && cfg.TLSCert != "" && cfg.TLSCACert != ""
}
