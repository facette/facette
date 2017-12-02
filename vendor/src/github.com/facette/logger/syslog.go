package logger

import (
	"fmt"
	"log/syslog"
)

var (
	syslogMap = map[int]syslog.Priority{
		LevelError:   syslog.LOG_ERR,
		LevelWarning: syslog.LOG_WARNING,
		LevelNotice:  syslog.LOG_NOTICE,
		LevelInfo:    syslog.LOG_INFO,
		LevelDebug:   syslog.LOG_DEBUG,
	}

	syslogFacilities = map[string]syslog.Priority{
		"kern":     syslog.LOG_KERN,
		"user":     syslog.LOG_USER,
		"mail":     syslog.LOG_MAIL,
		"daemon":   syslog.LOG_DAEMON,
		"auth":     syslog.LOG_AUTH,
		"syslog":   syslog.LOG_SYSLOG,
		"lpr":      syslog.LOG_LPR,
		"news":     syslog.LOG_NEWS,
		"uucp":     syslog.LOG_UUCP,
		"cron":     syslog.LOG_CRON,
		"authpriv": syslog.LOG_AUTHPRIV,
		"ftp":      syslog.LOG_FTP,
		"local0":   syslog.LOG_LOCAL0,
		"local1":   syslog.LOG_LOCAL1,
		"local2":   syslog.LOG_LOCAL2,
		"local3":   syslog.LOG_LOCAL3,
		"local4":   syslog.LOG_LOCAL4,
		"local5":   syslog.LOG_LOCAL5,
		"local6":   syslog.LOG_LOCAL6,
		"local7":   syslog.LOG_LOCAL7,
	}
)

type syslogBackend struct {
	logger *Logger
	writer *syslog.Writer
	level  int
}

func newSyslogBackend(config SyslogConfig, logger *Logger) (backend, error) {
	var (
		writer *syslog.Writer
		err    error
	)

	if config.Level == "" {
		config.Level = defaultLevel
	} else if _, ok := levelMap[config.Level]; !ok {
		return nil, ErrInvalidLevel
	}

	// Check for syslog facility
	facility, ok := syslogFacilities[config.Facility]
	if !ok {
		return nil, ErrInvalidFacility
	}

	if config.Address != "" {
		network := config.Transport
		if network == "" {
			network = "udp"
		}

		writer, err = syslog.Dial(network, config.Address, facility, config.Tag)
	} else {
		writer, err = syslog.New(facility, config.Tag)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to initialize syslog: %s", err)
	}

	return &syslogBackend{
		logger: logger,
		writer: writer,
		level:  levelMap[config.Level],
	}, nil
}

func (b syslogBackend) Close() {
	b.writer.Close()
}

func (b syslogBackend) Write(level int, mesg string) {
	if level > b.level {
		return
	}

	switch level {
	case LevelError:
		b.writer.Err(mesg)

	case LevelWarning:
		b.writer.Warning(mesg)

	case LevelNotice:
		b.writer.Notice(mesg)

	case LevelInfo:
		b.writer.Info(mesg)

	case LevelDebug:
		b.writer.Debug(mesg)
	}
}
