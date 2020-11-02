package logger

import (
	"fmt"
	"log"
	"strings"
)

type Default struct {
	level      Level
	components []string
}

func NewDefault() *Default {
	return &Default{
		level: Info,
	}
}

func (s *Default) SetLevel(level Level) {
	s.level = level
}

func (s *Default) Level() Level {
	return s.level
}

func (s *Default) SetComponents(components []string) {
	s.components = components
}

func (s *Default) Components() []string {
	return s.components
}

func (s *Default) Debug(component string, message string, a ...interface{}) {
	s.log(Debug, component, message, a...)
}

func (s *Default) Info(component string, message string, a ...interface{}) {
	s.log(Info, component, message, a...)
}

func (s *Default) Warning(component string, message string, a ...interface{}) {
	s.log(Warning, component, message, a...)
}

func (s *Default) Error(component string, message string, a ...interface{}) {
	s.log(Error, component, message, a...)
}

func (s *Default) Fatal(component string, message string, a ...interface{}) {
	s.log(Fatal, component, message, a...)
}

func (s *Default) Panic(component string, message string, a ...interface{}) {
	s.log(Panic, component, message, a...)
}

func (s *Default) log(level Level, component string, message string, a ...interface{}) {
	message = strings.TrimSpace(fmt.Sprintf("%s [%s] "+message, append([]interface{}{ level.String(), component}, a...)...))

	if level == Fatal {
		log.Fatal(message)
		return
	}

	if level == Panic {
		log.Panic(message)
		return
	}

	if level >= s.level && s.isComponentEnabled(component) {
		log.Printf(message)
	}
}

func (s *Default) isComponentEnabled(component string) bool {
	if len(s.components) == 0 {
		return true
	}
	for _, c := range s.components {
		if strings.EqualFold(component, c) {
			return true
		}
	}
	return false
}
