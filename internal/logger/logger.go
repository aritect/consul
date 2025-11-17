package logger

import (
	"log"
	"os"
)

type Logger struct {
	*log.Logger
}

func New() *Logger {
	return &Logger{
		Logger: log.New(
			os.Stderr,
			"info: ",
			log.Ldate|log.Ltime,
		),
	}
}

func (l *Logger) Info(format string, args ...interface{}) {
	l.SetPrefix("info: ")

	if args != nil {
		l.Printf(format, args...)
	} else {
		l.Print(format)
	}
}

func (l *Logger) Warning(format string, args ...interface{}) {
	l.SetPrefix("warning: ")

	if args != nil {
		l.Printf(format, args...)
	} else {
		l.Print(format)
	}
}

func (l *Logger) Error(format string, args ...interface{}) {
	l.SetPrefix("error: ")

	if args != nil {
		l.Printf(format, args...)
	} else {
		l.Print(format)
	}
}
