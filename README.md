Go-metrics Exporters
===

Export [go-metrics](https://github.com/rcrowley/go-metrics) to multiple upstreams with Emitter, and gracefully close it.

Features
---

* Gracefully report last metrics when shutting down
* Report to multiple upstreams simultaneously
* Builtin influx v1 and v2 support
* Implement `Emitter` to support in-house upstreams

Usage Examples
---

Report to stdout:

```go
reg := metrics.NewRegistry()
stdout := emitters.NewStdoutEmitter()
rep, err := exporters.NewReporter(reg, 1*time.Second, exporters.WithEmitters(stdout))
rep.Start()

// Close it gracefaully
rep.Close()
```

Report to stdout and influx v2:

```go
inf, _ := influx.NewV2Emitter(influxUrl, "bucket")
rep, err := exporters.NewReporter(reg, 1*time.Second, exporters.WithEmitters(stdout, inf))
rep.Start()
```

See test for more examples.

Alternatives
---

See [go-metrics#publishing-metrics](https://github.com/rcrowley/go-metrics#publishing-metrics) for alternatives.