package logger

const (
	LevelFatal logLevel = "FATAL"
	LevelError logLevel = "ERROR"
	LevelWarn  logLevel = "WARN"
	LevelInfo  logLevel = "INFO"

	FilenameDateFormat = "2006-01-02"
)

const (
	LogTypeError = iota
	LogTypeRequest
)

type (
	logLevel string
)

var (
	LogsPath string
)
