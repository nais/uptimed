# uptimed

[![CircleCI](https://circleci.com/gh/nais/uptimed.svg?style=svg)](https://circleci.com/gh/nais/uptimed)
[![Go Report Card](https://goreportcard.com/badge/github.com/nais/uptimed)](https://goreportcard.com/report/github.com/nais/uptimed)
[![License](http://img.shields.io/badge/license-mit-blue.svg?style=flat-square)](https://raw.githubusercontent.com/nais/uptimed/master/LICENSE)

Tool for checking the uptime of a given endpoint

## usage

```
$ curl http://<uptimed>/start?url=<url>&timeout=1800&interval=2
<monitor_id>
$ // do stuff
$ curl http://<uptimed>/stop/<monitor_id>
<uptime result>
```


