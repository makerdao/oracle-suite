package logger

import (
	"errors"
	"strings"
)

type Level int

const (
	Invalid = iota - 2
	Debug
	Info
	Warning
	Error
	Fatal
	Panic
	None
)

func (l Level) String() string {
	switch l {
	case Info:
		return "INFO"
	case Debug:
		return "DEBUG"
	case Warning:
		return "WARN"
	case Error:
		return "ERROR"
	case Fatal:
		return "FATAL"
	case Panic:
		return "PANIC"
	case None:
		return "NONE"
	}

	return "INVALID"
}

func LevelFromString(level string) (Level, error) {
	switch strings.ToLower(level) {
	case "info":
		return Info, nil
	case "debug":
		return Debug, nil
	case "warn", "warning":
		return Warning, nil
	case "err", "error":
		return Error, nil
	case "fatal":
		return Fatal, nil
	case "panic":
		return Panic, nil
	case "none", "nothing":
		return None, nil
	}

	return Invalid, errors.New("invalid error level, valid levels are: info, debug, warning, error, fatal, panic and none")
}

type Logger interface {
	SetLevel(level Level)
	Level() Level
	SetTags(tags []string)
	Tags() []string
	Debug(tag string, message string, a ...interface{})
	Info(tag string, message string, a ...interface{})
	Warning(tag string, message string, a ...interface{})
	Error(tag string, message string, a ...interface{})
	Fatal(tag string, message string, a ...interface{})
	Panic(tag string, message string, a ...interface{})
}
