package config

import "github.com/alecthomas/kong"

// Cli holds command line args, flags and cmds
type Cli struct {
	Version  kong.VersionFlag
	Cfgfile  string `kong:"name='config',env='CONFIG',help='FTPGrab configuration file.'"`
	Schedule string `kong:"name='schedule',env='SCHEDULE',help='CRON expression format.'"`
	LogLevel string `kong:"name='log-level',env='LOG_LEVEL',default='info',help='Set log level.'"`
	LogJSON  bool   `kong:"name='log-json',env='LOG_JSON',default='false',help='Enable JSON logging output.'"`
	LogFile  string `kong:"name='log-file',env='LOG_FILE',help='Add logging to a specific file.'"`
}
