Go-metrics Exporters
===

[![GitHub Actions](https://img.shields.io/github/actions/workflow/status/juvenn/metric-exporters/build.yml?branch=master&style=flat-square)](https://github.com/juvenn/metric-exporters/actions)
[![GitHub Release](https://img.shields.io/github/release/juvenn/metric-exporters/all.svg?style=flat-square)](https://github.com/juvenn/metric-exporters/releases)
[![Go Report Card](https://goreportcard.com/badge/github.com/juvenn/metric-exporters?style=flat-square)](https://goreportcard.com/report/github.com/juvenn/metric-exporters)
![Go Version](https://img.shields.io/github/go-mod/go-version/juvenn/metric-exporters?style=flat-square)
[![License Apache-2.0](https://img.shields.io/github/license/juvenn/metric-exporters?style=flat-square)](https://github.com/juvenn/metric-exporters/blob/master/LICENSE)


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
rep, err := exporters.NewReporter(reg, 1*time.Second).WithEmitter(stdout).Start()

// Should close it when shutting down application
rep.Close()
```

Report to stdout and influx v2:

```go
inf, _ := influx.NewV2Emitter(influxUrl, "bucket")
rep, err := exporters.NewReporter(reg, 1*time.Second).
	WithEmitter(stdout).
	WithEmitter(inf).
	Start()
// Should close it when shutting down application
rep.Close()
```

See test for more examples.

Alternatives
---

See [go-metrics#publishing-metrics](https://github.com/rcrowley/go-metrics#publishing-metrics) for alternatives.