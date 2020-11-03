package logger

import (
	"fmt"
	"log"
	"strings"
)

type Default struct {
	level      Level
	tags []string
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

func (s *Default) SetTags(tags []string) {
	s.tags = tags
}

func (s *Default) Tags() []string {
	return s.tags
}

func (s *Default) Debug(tag string, message string, a ...interface{}) {
	s.log(Debug, tag, message, a...)
}

func (s *Default) Info(tag string, message string, a ...interface{}) {
	s.log(Info, tag, message, a...)
}

func (s *Default) Warning(tag string, message string, a ...interface{}) {
	s.log(Warning, tag, message, a...)
}

func (s *Default) Error(tag string, message string, a ...interface{}) {
	s.log(Error, tag, message, a...)
}

func (s *Default) Fatal(tag string, message string, a ...interface{}) {
	s.log(Fatal, tag, message, a...)
}

func (s *Default) Panic(tag string, message string, a ...interface{}) {
	s.log(Panic, tag, message, a...)
}

func (s *Default) log(level Level, tag string, message string, a ...interface{}) {
	message = strings.TrimSpace(fmt.Sprintf("%-5s [%s] "+message, append([]interface{}{ level.String(), tag}, a...)...))

	if level == Fatal {
		log.Fatal(message)
		return
	}

	if level == Panic {
		log.Panic(message)
		return
	}

	if level >= s.level && s.isTagEnabled(tag) {
		log.Printf(message)
	}
}

func (s *Default) isTagEnabled(tag string) bool {
	if len(s.tags) == 0 {
		return true
	}
	for _, c := range s.tags {
		if strings.EqualFold(tag, c) {
			return true
		}
	}
	return false
}