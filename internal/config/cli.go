package config

import "github.com/alecthomas/kong"

// Cli holds command line args, flags and cmds
type Cli struct {
	Version      kong.VersionFlag
	Cfgfile      string `kong:"name='config',env='CONFIG',help='FTPGrab configuration file.'"`
	Schedule     string `kong:"name='schedule',env='SCHEDULE',help='CRON expression format.'"`
	LogLevel     string `kong:"name='log-level',env='LOG_LEVEL',default='info',help='Set log level.'"`
	LogJSON      bool   `kong:"name='log-json',env='LOG_JSON',default='false',help='Enable JSON logging output.'"`
	LogTimestamp bool   `kong:"name='log-timestamp',env='LOG_TIMESTAMP',default='true',help='Adds the current local time as UNIX timestamp to the logger context.'"`
	LogCaller    bool   `kong:"name='log-caller',env='LOG_CALLER',default='false',help='Add file:line of the caller to log output.'"`
	LogFile      string `kong:"name='log-file',env='LOG_FILE',help='Add logging to a specific file.'"`
}
