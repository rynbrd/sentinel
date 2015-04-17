Sentinel
========
Triggered templating and command execution for etcd.

[![Build Status](https://travis-ci.org/BlueDragonX/sentinel.svg?branch=master)](https://travis-ci.org/BlueDragonX/sentinel)

How It Works
------------
Sentinel allows you to trigger template generation and command execution off of
key changes in etcd. Sentinel defines a `watcher` to wait on one or more
`watch` keys. The watcher then defines `templates`, `context`, and a `command`
to execute when a `watch` key changes. The `command` is executed only when one
or more of the destination templates files changed. If no templates are
provided to a watcher then the command will always be executed. A watcher may
also be defined with one ore more templates and no command. The `context` is is
a list of keys which are retrieved from etcd and passed into each template for
rendering.

Installing
----------
This is a go-gettable package. It should be as simple as:

    go get github.com/BlueDragonX/sentinel

Configuring
-----------
The `sentinel` binary takes two optional arguments: `-config` and `-exec`. The
`-config` argument takes as option the path to the Sentinel configuration file.
This defaults to `config.yml` in the current directory. The `-exec` argument
causes Sentinel to execute a watcher's actions directly instead of running
continuously and waiting for its watchers. The `-exec` argument may be provided
multiple times to execute multiple watchers.

The config file is written in YAML. It is structure into four sections: `etcd`,
`watchers`, and `logging`.

### etcd ###
This section configures the connection to etcd. Available parameters are:

- `uri` - The URI to connect to etcd at. Defaults to `http://172.17.42.1:4001/`.
- `uris` - Connect to multiple etcd nodes. Used as an alternative to `uri` when
  redundancy is called for.
- `prefix` - All etcd key paths will be prefixed with this value. Defaults to
  an empty string.
- `tls-key` - The path to the TLS private key to use when connecting. Must be
  provided to enable TLS.
- `tls-cert` - The path to the TLS certificate to use when connecting. Must be
  provided to enable TLS.
- `tls-ca-cert` - The path to the TLS CA certificate to use when connecting.
  Must be provided to enable TLS.

### watchers ###
This section defines watchers to trigger off of etcd key changes. The watchers
section is a mapping of watcher names to their configuration. Available watcher
paremeters are:

- `prefix` - Key paths will be prefixed with this value. This will have the
  `etcd.prefix` value prepended to it. Defaults to an empty string. This allows
  you to reuse a template across multiple watchers whose keys would otherwise
  look the same.
- `watch` - A list of etcd keys to wait watch for changes. These are
  automatically prefixed with the the value of `etcd.prefix`.
- `context` - A list of keys whose values will be retrieved and passed to the
  templates to render. These values are retrieved recursively. Key values will
  have dashes (`-`) replaces with underscored (`_`) so as to be accessible in
  the templates.
- `templates` - A list of templates to render. Each template is a mapping
  containing a `src` and `dest` value. The `src` is the template source code
  and the `dest` is the place where the rendered template will be written to.
  Directories under `dest` will be created if necessary.
- `command` - The command to execute. If templates are provided then this
  command will be only be executed when one or more template destinations are
  changed. The command may be one of two forms: a string or an array of
  arguments. The first form will cause the command to be executed in a bash
  shell. The second will cause it to be executed directly.

### logging ###
This section controls how Beacon outputs logging. Sentinel uses [go-log][3] for
logging. See its documentation for valid target and log level values.

- `target` - The target to log to. Defaults to `stderr`.
- `level` - The log level. Valid values are `debug`, `info`, or `error`.

Template Functions
------------------
A handful of template functions have been added to make configuring certain
things easier. These are:

- `addrHost` - Return the host part of a host:port formatted address.
- `addrPort` - Return the port part of a host:port formatted address.
- `urlScheme` - Return the scheme part of a URL.
- `urlHost` - Return the host part of a URL. This include the :port if present.
- `urlUsername` - Return the username part of a URL.
- `urlPassword` - Return the password part of a URL.
- `urlRawQuery` - Return the URL's query string.
- `urlQuery` - Return the first value of a query key. Takes `name` as an additional parameter.
- `urlFragment` - Return the fragment part of the URL.
- `json` - Unmarshal a value into a JSON map or array.

All functions return an empty string on error.

Beacon Example
--------------
[Beacon][2] discovers services running in Docker and registers them in etcd.
The following Sentinel configuration maintains a list of registry endpoints:

    etcd:
      uri: http://localhost:4001
      prefix: beacon
    
    watchers:
      registries:
        watch:
        - registry/_index
        context:
        - registry
        templates:
        - src: /tmp/registries.tpl
          dest: /tmp/registries.yml

The registries.tpl template looks like this:

    registries:
    {{range $containerId, $container := .registry}}- http://{{$container.host_name}}:{{$container.host_port}}/{{end}}

Assuming there is one container listening on port `2002` whose hostname is
`docker.example.net` the output of `sentinel -exec registries` would look like this:

    registries:
    - http://docker.example.net:2002/

License
-------
Copyright (c) 2014 Ryan Bourgeois. Licensed under BSD-Modified. See the
[LICENSE][1] file for a copy of the license.

[1]: https://raw.githubusercontent.com/BlueDragonX/sentinel/master/LICENSE "Sentinel License"
[2]: https://github.com/BlueDragonX/beacon/ "Beacon"
[3]: https://github.com/BlueDragonX/go-log/ "go-log"
