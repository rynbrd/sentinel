package main

import "testing"

func TestAddrHost(t *testing.T) {
	tests := [][2]string{
		{"", ""},
		{":http", ""},
		{"localhost", "localhost"},
		{"localhost:http", "localhost"},
	}

	for _, test := range tests {
		have := AddrHost(test[0])
		if have != test[1] {
			t.Errorf("%s != %s", have, test[1])
		}
	}
}

func TestAddrPort(t *testing.T) {
	tests := [][2]string{
		{"", ""},
		{":http", "http"},
		{"localhave", ""},
		{"localhave:http", "http"},
		{":80", "80"},
		{"localhave:80", "80"},
	}

	for _, test := range tests {
		have := AddrPort(test[0])
		if have != test[1] {
			t.Errorf("%s != %s", have, test[1])
		}
	}
}

func TestURLScheme(t *testing.T) {
	tests := [][2]string{
		{"", ""},
		{"not a URL", ""},
		{"://localhost/", ""},
		{"http://localhost/", "http"},
		{"file:///proc/self", "file"},
	}

	for _, test := range tests {
		have := URLScheme(test[0])
		if have != test[1] {
			t.Errorf("%s != %s", have, test[1])
		}
	}
}

func TestURLUsername(t *testing.T) {
	tests := [][2]string{
		{"", ""},
		{"not a URL", ""},
		{"http://user@localhost/", "user"},
		{"http://user:pass@localhost/", "user"},
		{"http://:pass@localhost/", ""},
		{"file:///proc/self", ""},
	}

	for _, test := range tests {
		have := URLUsername(test[0])
		if have != test[1] {
			t.Errorf("%s != %s", have, test[1])
		}
	}
}

func TestURLPassword(t *testing.T) {
	tests := [][2]string{
		{"", ""},
		{"not a URL", ""},
		{"http://user:pass@localhost/", "pass"},
		{"http://:pass@localhost/", "pass"},
		{"http://user:@localhost/", ""},
		{"http://user@localhost/", ""},
		{"file:///proc/self", ""},
	}

	for _, test := range tests {
		have := URLPassword(test[0])
		if have != test[1] {
			t.Errorf("%s != %s", have, test[1])
		}
	}
}

func TestURLHost(t *testing.T) {
	tests := [][2]string{
		{"", ""},
		{"not a URL", ""},
		{"http:///", ""},
		{"file:///proc/self", ""},
		{"http://localhost/", "localhost"},
		{"http://localhost:8080/", "localhost:8080"},
		{"http://user@localhost/", "localhost"},
	}

	for _, test := range tests {
		have := URLHost(test[0])
		if have != test[1] {
			t.Errorf("%s != %s", have, test[1])
		}
	}
}

func TestURLPath(t *testing.T) {
	tests := [][2]string{
		{"", ""},
		{"not a URL", "not a URL"},
		{"http:///", "/"},
		{"file:///proc/self", "/proc/self"},
		{"http://localhost/", "/"},
		{"http://localhost", ""},
	}

	for _, test := range tests {
		have := URLPath(test[0])
		if have != test[1] {
			t.Errorf("%s != %s", have, test[1])
		}
	}
}

func TestURLRawQuery(t *testing.T) {
	tests := [][2]string{
		{"", ""},
		{"not a URL", ""},
		{"http:///", ""},
		{"http:///?query", "query"},
		{"http://localhost/?query", "query"},
		{"http://localhost/?name=value", "name=value"},
	}

	for _, test := range tests {
		have := URLRawQuery(test[0])
		if have != test[1] {
			t.Errorf("%s != %s", have, test[1])
		}
	}
}

func TestURLQuery(t *testing.T) {
	tests := [][3]string{
		{"", "name", ""},
		{"not a URL", "name", ""},
		{"http:///", "name", ""},
		{"http:///?query", "name", ""},
		{"http://localhost/?query", "name", ""},
		{"http://localhost/?name=value", "name", "value"},
		{"http://localhost/?name=value1,value2", "name", "value1,value2"},
		{"http://localhost/?name=value1&name=value2", "name", "value1"},
		{"http://localhost/?other=value&name=value1", "name", "value1"},
		{"http://localhost/?other=value&name=value1", "other", "value"},
	}

	for _, test := range tests {
		have := URLQuery(test[0], test[1])
		if have != test[2] {
			t.Errorf("%s != %s", have, test[2])
		}
	}
}

func TestURLFragment(t *testing.T) {
	tests := [][2]string{
		{"", ""},
		{"not a URL", ""},
		{"file://proc/self", ""},
		{"file://proc/self#pid", "pid"},
		{"http://localhost/", ""},
		{"http://localhost/#", ""},
		{"http://localhost/#frag", "frag"},
	}

	for _, test := range tests {
		have := URLFragment(test[0])
		if have != test[1] {
			t.Errorf("%s != %s", have, test[1])
		}
	}
}
