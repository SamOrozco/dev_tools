package logger

import "github.com/fatih/color"

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
	color.Red("%s [Error] - %s", s.prefix, message)
}

func (s StdOutLogger) Warn(message string) {
	color.HiRed("%s [Warn] - %s", s.prefix, message)
}

func (s StdOutLogger) Debug(message string) {
	color.Cyan("%s [Debug] - %s", s.prefix, message)
}
