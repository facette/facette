package logger

// FileConfig represents a file backend configuration.
type FileConfig struct {
	// Logging output severity level. Messages with higher severity value will be discarded.
	Level string

	// File path of the logging output. If path is either empty or "-", logging will be output to os.Stderr.
	Path string
}

// SyslogConfig represents a syslog backend configuration.
type SyslogConfig struct {
	// Logging output severity level. Messages with higher severity value will be discarded.
	Level string

	// syslog facility to send messages to.
	Facility string

	// syslog tag to specify in messages.
	Tag string

	// syslog service address and transport type (either "udp", "tcp" or "unix"). If not sepcified, local syslog will
	// be used.
	Address   string
	Transport string
}
