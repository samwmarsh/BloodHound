package environment

import "strings"

type envVar string

const (
	Prefix = "SB_"
)

const (
	LogLevel envVar = "LOG_LEVEL"
	Verbose  envVar = "VERBOSE"
)

func (ev envVar) Env() string {
	return Prefix + string(ev)
}

func (ev envVar) Flag() string {
	return strings.ToLower(string(ev))
}
