package logger

import (
	"fmt"
	"os"
	"time"
)

type Level = string

const (
	DEBUG Level = "DEBUG"
	WARN  Level = "FAIL"
	INFO  Level = "INFO"
	ERROR Level = "ERROR"
)

var levelColors = map[Level]string{
	DEBUG: "\033[36m", // Cyan
	WARN:  "\033[33m", // Yellow
	INFO:  "\033[32m", // Green
	ERROR: "\033[31m", // Red
}

const resetColor = "\033[0m"

type Logger struct {
	level Level
	order map[Level]int
}

func New(level Level) *Logger {
	return &Logger{
		level: level,
		order: map[Level]int{
			DEBUG: 0,
			WARN:  1,
			INFO:  2,
			ERROR: 3,
		},
	}
}

func (l *Logger) log(lvl Level, format string, args ...any) {
	if l.order[lvl] < l.order[l.level] {
		return
	}

	timestamp := time.Now().Format("2006-01-02 15:04:05")
	coloredLevel := fmt.Sprintf("%s%s%s", levelColors[lvl], lvl, resetColor)
	msg := fmt.Sprintf(format, args...)
	fmt.Fprintf(os.Stdout, "[%s] [%s] %s\n", timestamp, coloredLevel, msg)
}

func (l *Logger) Debug(format string, args ...any) {
	l.log(DEBUG, format, args...)
}

func (l *Logger) Warn(format string, args ...any) {
	l.log(WARN, format, args...)
}

func (l *Logger) Info(format string, args ...any) {
	l.log(INFO, format, args...)
}

func (l *Logger) Error(format string, args ...any) {
	l.log(ERROR, format, args...)
}
