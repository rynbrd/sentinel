package main

import (
	"errors"
	"gopkg.in/BlueDragonX/yamlcfg.v1"
)

const (
	DefaultLoggingTarget  = "stderr"
	DefaultLoggingLevel   = "info"
)

// Store log related configuration.
type LoggingConfig struct {
	Target string
	Level  string
}

// Get default logging config.
func DefaultLoggingConfig() LoggingConfig {
	return LoggingConfig{
		DefaultLoggingTarget,
		DefaultLoggingLevel,
	}
}

// SetYAML parses the YAML tree into the object.
func (cfg *LoggingConfig) SetYAML(tag string, data interface{}) bool {
	yamlcfg.AssertIsMap("logging", data)
	cfg.Target = yamlcfg.GetString(data, "target", DefaultLoggingTarget)
	cfg.Level = yamlcfg.GetString(data, "level", DefaultLoggingLevel)
	return true
}

// Validate the logging configuration.
func (cfg *LoggingConfig) Validate() []error {
	errs := []error{}
	if cfg.Target == "" {
		errs = append(errs, errors.New("invalid logging target"))
	}
	return errs
}
