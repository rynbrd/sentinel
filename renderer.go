package main

import (
	"fmt"
	"gopkg.in/BlueDragonX/go-hash.v1"
	"gopkg.in/BlueDragonX/simplelog.v1"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

type Template struct {
	Src    string
	Dest   string
	logger *simplelog.Logger
}

func NewTemplate(src, dest string, logger *simplelog.Logger) Template {
	return Template{src, dest, logger}
}

// Return true if one file differs from another.
func (t *Template) differs(fileA, fileB string) bool {
	var err error
	var hashA, hashB string
	if hashA, err = hash.File(fileA); err != nil {
		t.logger.Warn("unable to hash %s", fileA)
		return true
	}
	if hashB, err = hash.File(fileB); err != nil {
		t.logger.Warn("unable to hash %s", fileB)
		return true
	}
	return hashA != hashB
}

// Render the template to a temporary and return true if the original was changed.
func (t *Template) Render(context map[string]interface{}) (changed bool, err error) {
	// create the destination directory
	dir := filepath.Dir(t.Dest)
	if err = os.MkdirAll(dir, 0777); err != nil {
		return
	}

	// create a temp file to write
	var tmp *os.File
	prefix := fmt.Sprintf(".%s-", filepath.Base(t.Dest))
	if tmp, err = ioutil.TempFile(dir, prefix); err != nil {
		return
	}
	defer func() {
		tmp.Close()
		if !changed || err != nil {
			os.Remove(tmp.Name())
		}
	}()

	// add functions to the templates
	funcs := template.FuncMap{
		"replace": strings.Replace,
	}

	// render the template to the temp file
	var tpl *template.Template
	name := filepath.Base(t.Src)
	if tpl, err = template.New(name).Funcs(funcs).ParseFiles(t.Src); err != nil {
		return
	}
	if err = tpl.Execute(tmp, context); err != nil {
		return
	}
	tmp.Close()

	// return if the old and new files are the same
	changed = t.differs(t.Dest, tmp.Name())
	if !changed {
		return
	}

	// replace the old file with the new one
	err = os.Rename(tmp.Name(), t.Dest)
	return
}

// A renderer generates files from a collection of templates.
type Renderer struct {
	templates []Template
	logger    *simplelog.Logger
}

func NewRenderer(templates []Template, logger *simplelog.Logger) *Renderer {
	var item *Renderer

	if len(templates) == 0 {
		return item
	}

	item = &Renderer{
		templates,
		logger,
	}
	return item
}

func (renderer *Renderer) Render(context map[string]interface{}) (changed bool, err error) {
	var oneChanged bool
	for _, template := range renderer.templates {
		if oneChanged, err = template.Render(context); err != nil {
			return
		}
		if oneChanged {
			renderer.logger.Debug("template '%s' rendered to '%s'", template.Src, template.Dest)
		} else {
			renderer.logger.Debug("template '%s' did not change", template.Dest)
		}
		changed = changed || oneChanged
	}
	return changed, nil
}
