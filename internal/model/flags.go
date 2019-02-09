package model

// Flags holds flags from command line
type Flags struct {
	Cfgfile    string
	Output     string
	Schedule   string
	Timezone   string
	LogLevel   string
	LogNocolor bool
	LogFile    bool
	LogFtp     bool
}
