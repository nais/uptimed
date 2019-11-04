# uptimed

[![CircleCI](https://circleci.com/gh/nais/uptimed.svg?style=svg)](https://circleci.com/gh/nais/uptimed)
[![Go Report Card](https://goreportcard.com/badge/github.com/nais/uptimed)](https://goreportcard.com/report/github.com/nais/uptimed)
[![License](http://img.shields.io/badge/license-mit-blue.svg?style=flat-square)](https://raw.githubusercontent.com/nais/uptimed/master/LICENSE)

Tool for checking the uptime of a given endpoint

## Usage

```
$ curl -X POST http://<uptimed>/start?endpoint=<url>&timeout=1800&interval=2
<monitor_id>
$ // do stuff
$ curl -X POST http://<uptimed>/stop/<monitor_id>
<uptime result>
```
## Development

[Go modules](https://github.com/golang/go/wiki/Modules)
are used for dependency tracking. Make sure you do `export GO111MODULE=on` before running any Go commands.
It is no longer needed to have the project checked out in your `$GOPATH`.

```
go run uptimed.go --bind-address=127.0.0.1:8080
```

