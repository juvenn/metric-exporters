go-metrics reporter
===

Report [go-metrics](https://github.com/rcrowley/go-metrics) to multiple upstreams with Emitter, and gracefully close it, i.e. wait and report last metrics before exit.

Usage Example
---

Report to stdout and influx v2:

```go
reg := metrics.NewRegistry()
stdout := emitters.NewStdoutEmitter()
inf, _ := influx.NewV2Emitter(influxUrl, "bucket")
rep, err := exporters.NewReporter(reg, 1*time.Second, exporters.WithEmitters(stdout, inf))
rep.Start()

// Close it gracefaully
rep.Close()
```

You can define your own emitter by implementing `Emitter` interface.