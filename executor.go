package main

import (
	"os/exec"
	"strings"
)

// An Executor is responsible for executed to perform some action when a
// watched key is changed.
type Executor interface {
	// Return the unique name of the executor.
	Name() string

	// Called to run the executor's actions.
	Execute(client Client) error
}

// A Executor performs template rendering. It will optionally execute a command
// when one or more changes are made by the templating system. If no templates
// are provided the command will always be executed.
type TemplateExecutor struct {
	name      string
	prefix    string
	context   []string
	Templates []Template
	Command   []string
}

// Render the templates. Return true if any templates changed.
func (ex *TemplateExecutor) render(context interface{}) (changed bool, err error) {
	var oneChanged bool
	if ex.Templates == nil || len(ex.Templates) == 0 {
		logger.Debugf("%s: no templates to render", ex.name)
		changed = true
		return
	}

	logger.Debugf("%s: context %+v", ex.name, context)
	for _, tpl := range ex.Templates {
		if oneChanged, err = tpl.Render(context); err != nil {
			return
		}
		if oneChanged {
			logger.Debugf("%s: rendered '%s' -> '%s'", ex.name, tpl.Src, tpl.Dest)
		} else {
			logger.Debugf("%s: no change to '%s'", ex.name, tpl.Dest)
		}
		changed = changed || oneChanged
	}
	return
}

// Run the command.
func (ex *TemplateExecutor) run() error {
	if len(ex.Command) == 0 {
		logger.Debugf("%s: no command to call", ex.name)
		return nil
	}

	cmdName := ex.Command[0]
	cmdArgs := ex.Command[1:]
	command := exec.Command(cmdName, cmdArgs...)

	out, err := command.CombinedOutput()
	if err == nil {
		logger.Debugf("%s: command %v ran", ex.name, ex.Command)
	} else {
		logger.Errorf("%s: command %v failed: %s", ex.name, ex.Command, err)
	}
	outStr := string(out)
	if outStr != "" {
		lines := strings.Split(outStr, "\n")
		for _, line := range lines {
			if err == nil {
				logger.Debugf("%s> %s", ex.name, line)
			} else {
				logger.Errorf("%s> %s", ex.name, line)
			}
		}
	}
	return err
}

// Return the unique name of the executor.
func (ex *TemplateExecutor) Name() string {
	return ex.name
}

// Render the templates using the context retrieved from the provided `client`
// and execute the command. The command will be executed if one of the template
// destinations changes or no templates are present in the Watcher.
func (ex *TemplateExecutor) Execute(client Client) error {
	var err error
	var context interface{}

	logger.Debugf("%s: executing", ex.name)
	if ex.context == nil || len(ex.context) == 0 {
		context = map[string]interface{}{}
	} else if context, err = client.Get(ex.context); err != nil {
		logger.Errorf("%s: context get failed: %s", ex.name, err)
		return err
	} else {
		for _, key := range strings.Split(ex.prefix, "/") {
			if contextMap, ok := context.(map[string]interface{}); ok {
				context, ok = contextMap[getKeyName(key)]
				if !ok {
					context = map[string]interface{}{}
					break
				}
			} else {
				context = map[string]interface{}{}
				break
			}
		}
	}

	run := true
	run, err = ex.render(context)
	if run && err == nil {
		err = ex.run()
	}
	return err
}
