# uptimed

tool for checking the uptime of a given endpoint

## usage

```
$ curl http://<uptimed>/start?url=<url>&timeout=1800&interval=2
<monitor_id>
$ // do stuff
$ curl http://<uptimed>/stop/<monitor_id>
<uptime result>
```


