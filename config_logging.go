package main

import (
	"gopkg.in/BlueDragonX/simplelog.v1"
	"gopkg.in/BlueDragonX/yamlcfg.v1"
)

const (
	DefaultLoggingSyslog  = false
	DefaultLoggingConsole = true
	DefaultLoggingLevel   = simplelog.NOTICE
)

// Store log related configuration.
type LoggingConfig struct {
	Syslog  bool
	Console bool
	Level   int
}

// Get default logging config.
func DefaultLoggingConfig() LoggingConfig {
	return LoggingConfig{
		DefaultLoggingSyslog,
		DefaultLoggingConsole,
		DefaultLoggingLevel,
	}
}

// SetYAML parses the YAML tree into the object.
func (cfg *LoggingConfig) SetYAML(tag string, data interface{}) bool {
	yamlcfg.AssertIsMap("logging", data)
	cfg.Syslog = yamlcfg.GetBool(data, "syslog", DefaultLoggingSyslog)
	cfg.Console = yamlcfg.GetBool(data, "console", DefaultLoggingConsole)
	levelStr := yamlcfg.GetString(data, "level", "")
	if levelStr == "" {
		cfg.Level = DefaultLoggingLevel
	} else {
		cfg.Level = simplelog.StringToLevel(levelStr)
	}
	return true
}

// Validate the logging configuration.
func (cfg *LoggingConfig) Validate() []error {
	return []error{}
}
