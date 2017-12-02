# logger: logging handler

Logging handler for Go.

[![GoDoc](https://godoc.org/github.com/facette/logger?status.svg)](https://godoc.org/github.com/facette/logger)

## Features

 * File and syslog logging backend
 * Colors when printing to `stderr`
 * Multiple logging backend support
 * Logging contexts

## Example

The following code:
```go
logger, err := logger.NewLogger(
	logger.FileConfig{
		Level: "debug",
		Path:  "/path/to/file.log",
	},
	logger.SyslogConfig{
		Tag:      "myapp",
		Level:    "warning",
		Facility: "local7",
	},
)
if err != nil {
	log.Fatalf("failed to initialize logger: %s", err)
}

logger.Info("begin")

ctx := logger.Context("test")
ctx.Info("entering context")
ctx.Debug("start time: %s", time.Now())
ctx.Info("leaving context")

logger.Warning("this is a sample warning")

logger.Info("end")

```

will output in `/path/to/file.log`:

```
2016/08/21 12:39:23.093357 INFO: begin
2016/08/21 12:39:23.093438 INFO: test: entering context
2016/08/21 12:39:23.093458 DEBUG: test: start time: 2016-08-21 12:39:23.093447326 +0200 CEST
2016/08/21 12:39:23.093475 INFO: test: leaving context
2016/08/21 12:39:23.093480 WARNING: this is a sample warning
2016/08/21 12:39:23.093485 INFO: end
```

and only sends "this is a sample warning" to syslog on _local7_ facility.
