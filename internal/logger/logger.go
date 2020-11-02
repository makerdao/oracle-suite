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
)

func (l Level) String() string {
	switch l {
	case Info:
		return "Info"
	case Debug:
		return "Debug"
	case Warning:
		return "Warning"
	case Error:
		return "Error"
	case Fatal:
		return "Fatal"
	case Panic:
		return "Panic"
	}

	return "Invalid"
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
	}

	return Invalid, errors.New("invalid error level, valid levels are: info, debug, warning, error, fatal and panic")
}

type Logger interface {
	SetLevel(level Level)
	Level() Level
	SetComponents(components []string)
	Components() []string
	Debug(component string, message string, a ...interface{})
	Info(component string, message string, a ...interface{})
	Warning(component string, message string, a ...interface{})
	Error(component string, message string, a ...interface{})
	Fatal(component string, message string, a ...interface{})
	Panic(component string, message string, a ...interface{})
}
