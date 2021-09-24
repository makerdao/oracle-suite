package callback

import (
	"fmt"

	"github.com/makerdao/oracle-suite/pkg/log"
)

type CallbackFunc func(level log.Level, fields log.Fields, log string)

// Logger implements the log.Logger interface. It allows using a custom
// callback function that will be invoked every time a log message is created.
type Logger struct {
	level    log.Level
	fields   log.Fields
	callback CallbackFunc
}

func New(level log.Level, callback CallbackFunc) *Logger {
	return &Logger{
		level:    level,
		fields:   log.Fields{},
		callback: callback,
	}
}

func (c *Logger) Level() log.Level {
	return c.level
}

func (c *Logger) WithField(key string, value interface{}) log.Logger {
	f := log.Fields{}
	for k, v := range c.fields {
		f[k] = v
	}
	f[key] = value
	return &Logger{
		level:    c.level,
		fields:   f,
		callback: c.callback,
	}
}

func (c *Logger) WithFields(fields log.Fields) log.Logger {
	f := log.Fields{}
	for k, v := range c.fields {
		f[k] = v
	}
	for k, v := range fields {
		f[k] = v
	}
	return &Logger{
		level:    c.level,
		fields:   f,
		callback: c.callback,
	}
}

func (c *Logger) WithError(err error) log.Logger {
	return c.WithField("err", err.Error())
}

func (c *Logger) Debugf(format string, args ...interface{}) {
	if c.level >= log.Debug {
		c.callback(c.level, c.fields, fmt.Sprintf(format, args...))
	}
}

func (c *Logger) Infof(format string, args ...interface{}) {
	if c.level >= log.Info {
		c.callback(c.level, c.fields, fmt.Sprintf(format, args...))
	}
}

func (c *Logger) Warnf(format string, args ...interface{}) {
	if c.level >= log.Warn {
		c.callback(c.level, c.fields, fmt.Sprintf(format, args...))
	}
}

func (c *Logger) Errorf(format string, args ...interface{}) {
	if c.level >= log.Error {
		c.callback(c.level, c.fields, fmt.Sprintf(format, args...))
	}
}

func (c *Logger) Panicf(format string, args ...interface{}) {
	c.callback(c.level, c.fields, fmt.Sprintf(format, args...))
}

func (c *Logger) Debug(args ...interface{}) {
	if c.level >= log.Debug {
		c.callback(c.level, c.fields, fmt.Sprint(args...))
	}
}

func (c *Logger) Info(args ...interface{}) {
	if c.level >= log.Info {
		c.callback(c.level, c.fields, fmt.Sprint(args...))
	}
}

func (c *Logger) Warn(args ...interface{}) {
	if c.level >= log.Warn {
		c.callback(c.level, c.fields, fmt.Sprint(args...))
	}
}

func (c *Logger) Error(args ...interface{}) {
	if c.level >= log.Error {
		c.callback(c.level, c.fields, fmt.Sprint(args...))
	}
}

func (c *Logger) Panic(args ...interface{}) {
	c.callback(c.level, c.fields, fmt.Sprint(args...))
}
