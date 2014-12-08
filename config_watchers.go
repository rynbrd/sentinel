package main

import (
	"errors"
	"fmt"
	"gopkg.in/BlueDragonX/simplelog.v1"
	"gopkg.in/BlueDragonX/yamlcfg.v1"
)

// Watcher template configuration.
type TemplateConfig struct {
	Src  string
	Dest string
}

// Parse the YAML tree into the object.
func (cfg *TemplateConfig) SetYAML(tag string, data interface{}) bool {
	yamlcfg.AssertIsMap(tag, data)
	cfg.Src = yamlcfg.GetString(data, "src", "")
	cfg.Dest = yamlcfg.GetString(data, "dest", "")
	return true
}

// Validate the file config object.
func (cfg *TemplateConfig) Validate() []error {
	errs := []error{}
	if cfg.Src == "" {
		errs = append(errs, errors.New("invalid value for template.src"))
	} else if !fileIsReadable(cfg.Src) {
		errs = append(errs, errors.New("invalid value for template.src: file is not readable"))
	}
	if cfg.Dest == "" {
		errs = append(errs, errors.New("invalid value for template.dest"))
	}
	return errs
}

// An array of templates.
type TemplatesConfig []TemplateConfig

// Parse the YAML tree into the object.
func (cfg *TemplatesConfig) SetYAML(tag string, data interface{}) bool {
	yamlcfg.AssertIsArray(tag, data)

	templates := []TemplateConfig{}
	for n, templateData := range data.([]interface{}) {
		template := TemplateConfig{}
		template.SetYAML(fmt.Sprintf("templates[%d]", n), templateData)
		templates = append(templates, template)
	}
	*cfg = templates
	return true
}

// Validate the file config object.
func (cfg *TemplatesConfig) Validate() []error {
	errs := []error{}
	for _, template := range *cfg {
		errs = append(errs, template.Validate()...)
	}
	return errs
}

// Watcher configuration.
type WatcherConfig struct {
	Name      string
	Prefix    string
	Watch     []string
	Context   []string
	Templates TemplatesConfig
	Cmd       []string
	Shell     bool
}

// Create a watcher from this config object.
func (cfg *WatcherConfig) CreateWatcher(client *Client, logger *simplelog.Logger) *Watcher {
	// create renderer
	templates := []Template{}
	for _, templateCfg := range cfg.Templates {
		templates = append(templates, NewTemplate(templateCfg.Src, templateCfg.Dest, logger))
	}
	var renderer *Renderer
	if len(templates) > 0 {
		renderer = NewRenderer(templates, logger)
	}

	// create command
	var command []string
	if len(cfg.Cmd) > 0 {
		if cfg.Shell {
			command = []string{"bash", "-c", cfg.Cmd[0]}
		} else {
			command = cfg.Cmd
		}
	}

	// create watcher
	return NewWatcher(cfg.Name, cfg.Prefix, cfg.Watch, cfg.Context, renderer, command, client, logger)
}

// Parse the YAML tree into the object.
func (cfg *WatcherConfig) SetYAML(tag string, data interface{}) bool {
	yamlcfg.AssertIsMap(tag, data)
	cfg.Name = tag
	cfg.Prefix = yamlcfg.GetString(data, "prefix", "")
	cfg.Watch = yamlcfg.GetStringArray(data, "watch", []string{})
	cfg.Context = yamlcfg.GetStringArray(data, "context", []string{})

	if templatesData, ok := yamlcfg.GetMapItem(data, "templates"); ok {
		cfg.Templates.SetYAML("templates", templatesData)
	}

	cfg.Shell = false
	if cmdValue, ok := yamlcfg.GetMapItem(data, "command"); ok {
		if _, ok := cmdValue.([]interface{}); ok {
			cfg.Cmd = yamlcfg.GetStringArray(data, "command", []string{})
		} else {
			shellCmd := yamlcfg.GetString(data, "command", "")
			if shellCmd == "" {
				cfg.Cmd = []string{}
			} else {
				cfg.Cmd = []string{shellCmd}
				cfg.Shell = true
			}
		}
	}
	return true
}

// Validate the file config object.
func (cfg *WatcherConfig) Validate() []error {
	errs := []error{}
	if len(cfg.Watch) == 0 {
		errs = append(errs, errors.New("invalid value for watcher.watch: no keys defined"))
	}
	if len(cfg.Context) == 0 && len(cfg.Templates) != 0 {
		errs = append(errs, errors.New("invalid value for watcher.context: templates require context keys"))
	}
	errs = append(errs, cfg.Templates.Validate()...)
	return errs
}

// An array of watchers.
type WatchersConfig []WatcherConfig

// Return the default watchers config.
func DefaultWatchersConfig() WatchersConfig {
	return WatchersConfig{}
}

// Parse the YAML tree into the object.
func (cfg *WatchersConfig) SetYAML(tag string, data interface{}) bool {
	yamlcfg.AssertIsMap(tag, data)

	watchers := []WatcherConfig{}
	for watcherName, watcherData := range data.(map[interface{}]interface{}) {
		yamlcfg.AssertIsString("watcher", watcherName)
		watcher := WatcherConfig{}
		watcher.SetYAML(watcherName.(string), watcherData)
		watchers = append(watchers, watcher)
	}
	*cfg = watchers
	return true
}

// Validate the file config object.
func (cfg *WatchersConfig) Validate() []error {
	errs := []error{}
	for _, watcher := range *cfg {
		errs = append(errs, watcher.Validate()...)
	}
	return errs
}

// Create a new watch manager from the configuration.
func (cfg *WatchersConfig) CreateWatchManager(client *Client, logger *simplelog.Logger) (manager *WatchManager, err error) {
	watchers := []*Watcher{}
	for _, watcherCfg := range *cfg {
		watchers = append(watchers, watcherCfg.CreateWatcher(client, logger))
	}
	return NewWatchManager(watchers, client, logger), nil
}
