package main

import (
	"io/ioutil"
	"os"
	"path"
	"reflect"
	"testing"
)

func TestTemplateDiffers(t *testing.T) {
	dir, err := ioutil.TempDir("", "sentinel_test_")
	if err != nil {
		t.Fatal("failed to create tempdir")
	}
	defer os.RemoveAll(dir)
	tpl := Template{}

	// only one file exists
	nameA := path.Join(dir, "only_one_a")
	nameB := path.Join(dir, "only_one_b")
	ioutil.WriteFile(nameA, []byte("example"), 0600)
	if !tpl.differs(nameA, nameB) {
		t.Error("differs returns incorrect value on missing file")
	}

	// files are the same
	nameA = path.Join(dir, "same_a")
	nameB = path.Join(dir, "same_b")
	value := []byte("example")
	ioutil.WriteFile(nameA, value, 0600)
	ioutil.WriteFile(nameB, value, 0600)
	if tpl.differs(nameA, nameB) {
		t.Error("differs returns incorrect value on identical files")
		fileA, _ := ioutil.ReadFile(nameA)
		fileB, _ := ioutil.ReadFile(nameB)
		t.Errorf("file a: %s", fileA)
		t.Errorf("file b: %s", fileB)
	}

	// flies are different
	nameA = path.Join(dir, "same_a")
	nameB = path.Join(dir, "same_b")
	ioutil.WriteFile(nameA, []byte("example_a"), 0600)
	ioutil.WriteFile(nameB, []byte("example_b"), 0600)
	if !tpl.differs(nameA, nameB) {
		t.Error("differs returns incorrect value on differing files")
	}
}

func TestTemplateRender(t *testing.T) {
	dir, err := ioutil.TempDir("", "sentinel_test_")
	if err != nil {
		t.Fatal("failed to create tempdir")
	}
	defer os.RemoveAll(dir)

	src := path.Join(dir, "src")
	dest := path.Join(dir, "dest")
	tpl := Template{Src: src, Dest: dest}
	ioutil.WriteFile(src, []byte("example: {{.example.name}}\n"), 0600)
	want := []byte("example: test\n")
	context := map[string]interface{}{
		"example": map[string]interface{}{"name": "test"},
	}

	// no out file
	if changed, err := tpl.Render(context); err == nil {
		if !changed {
			t.Error("template not written when dest is missing")
		}
		file, _ := ioutil.ReadFile(dest)
		if !reflect.DeepEqual(want, file) {
			t.Error("template dest is incorrect")
		}
	} else {
		t.Error(err)
	}

	// same out file
	if changed, err := tpl.Render(context); err == nil {
		if changed {
			t.Error("template written when dest is not changed")
		}
	} else {
		t.Error(err)
	}

	// out file differs
	ioutil.WriteFile(dest, []byte("these are not the droids you are looking for"), 0600)
	if changed, err := tpl.Render(context); err == nil {
		if !changed {
			t.Error("template not written when dest differs")
		}
		file, _ := ioutil.ReadFile(dest)
		if !reflect.DeepEqual(want, file) {
			t.Error("template dest is incorrect")
		}
	} else {
		t.Error(err)
	}
}
