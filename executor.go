package main

import (
	"fmt"
	"os/exec"
	"strings"
)

// An Executor is responsible for performing template rendering and command
// execution for a single watcher.
type Executor struct {
	Name      string
	Prefix    string
	Context   []string
	Templates []Template
	Command   []string
}

// Render the templates. Return true if any templates changed.
func (ex *Executor) render(context map[string]interface{}) (changed bool, err error) {
	var oneChanged bool
	for _, tpl := range ex.Templates {
		if oneChanged, err = tpl.Render(context); err != nil {
			return
		}
		if oneChanged {
			logger.Debugf("%s: rendered '%s' -> '%s'", ex.Name, tpl.Src, tpl.Dest)
		} else {
			logger.Debugf("%s: no change to '%s'", ex.Name, tpl.Dest)
		}
		changed = changed || oneChanged
	}
	return changed, nil
}

// Run the command.
func (ex *Executor) run() error {
	if len(ex.Command) == 0 {
		logger.Debugf("%s: command not set", ex.Name)
		return nil
	}

	cmdName := ex.Command[0]
	cmdArgs := ex.Command[1:]
	command := exec.Command(cmdName, cmdArgs...)

	out, err := command.CombinedOutput()
	if err == nil {
		logger.Debugf("%s: command ran", ex.Name)
	} else {
		logger.Errorf("%s: command failed: %s", ex.Name, err)
	}
	outStr := string(out)
	if outStr != "" {
		lines := strings.Split(outStr, "\n")
		for _, line := range lines {
			if err == nil {
				logger.Debugf("out: %s", line)
			} else {
				logger.Errorf("out: %s", line)
			}
		}
	}
	return err
}

// Render the templates using the context retrieved from the provided `client`
// and execute the command. The command will be executed if one of the template
// destinations changes or no templates are present in the Watcher.
func (ex *Executor) Execute(client *Client) error {
	var err error
	var context map[string]interface{}
	if context, err = client.Get(ex.Context); err != nil {
		logger.Errorf("%s: context failed: %s", ex.Name, err)
		return err
	}

	for _, key := range strings.Split(ex.Prefix, "/") {
		next, ok := context[key]
		if !ok {
			return fmt.Errorf("%s: context %s is invalid", ex.Name, ex.Prefix)
		}
		context, ok = next.(map[string]interface{})
		if !ok {
			return fmt.Errorf("%s: context %s is invalid", ex.Name, ex.Prefix)
		}
	}
	logger.Debugf("%s: got context: %v", ex.Name, context)

	run := true
	if len(ex.Templates) > 0 {
		run, err = ex.render(context)
	}
	logger.Debugf("run == %t", run)
	logger.Debugf("err == %v", err)
	logger.Debugf("cmd == %v", ex.Command)
	if len(ex.Command) > 0 && err == nil && run {
		err = ex.run()
	}
	return err
}
