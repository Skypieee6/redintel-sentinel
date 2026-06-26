package logger

import (
	"log"
	"os"
)

// Logger wraps the standard logger and serves as the foundation
// for future structured logging (JSON, levels, request IDs).
type Logger struct {
	*log.Logger
}

// New creates a logger that writes to stdout with date, time and source file.
func New() *Logger {
	return &Logger{
		Logger: log.New(os.Stdout, "[RedIntel] ", log.LstdFlags|log.Lshortfile),
	}
}

func (l *Logger) Info(msg string) {
	l.Println("INFO:", msg)
}

func (l *Logger) Error(msg string) {
	l.Println("ERROR:", msg)
}
