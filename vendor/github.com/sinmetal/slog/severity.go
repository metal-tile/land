package slog

// Severity is Stackdriver Logging Log Level
type Severity int

// Stackdriver Logging Log Level List
const (
	DEBUG Severity = iota
	INFO
	WARNING
	ERROR
)
