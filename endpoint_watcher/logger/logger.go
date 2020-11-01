package logger

import (
	"fmt"
	"github.com/fatih/color"
)

type Logger interface {
	Error(message string)
	Warn(message string)
	Debug(message string)
}

type StdOutLogger struct {
	prefix string
}

func NewStdOutLogger(pre string) Logger {
	return &StdOutLogger{
		prefix: pre,
	}
}

func (s StdOutLogger) Error(message string) {
	color.Red("%s [Error] - %s", getPrefix(s.prefix), message)
}

func (s StdOutLogger) Warn(message string) {
	color.HiRed("%s [Warn] - %s", getPrefix(s.prefix), message)
}

func (s StdOutLogger) Debug(message string) {
	color.Cyan("%s [Debug] - %s", getPrefix(s.prefix), message)
}

func getPrefix(pre string) string {
	if len(pre) < 1 {
		return ""
	}
	return fmt.Sprintf("[%s]", pre)
}
