package logger

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

type LogLevel int

const (
	Debug LogLevel = iota

	Info

	Warning

	Error
)

type LogEntry struct {
	Timestamp time.Time
	Level     LogLevel
	Message   string
}

func (l LogLevel) String() string {
	switch l {
	case Debug:
		return "DEBUG"
	case Info:
		return "INFO"
	case Warning:
		return "WARN"
	case Error:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}

var (
	instance *Logger
	once     sync.Once
)

type Logger struct {
	entries     []LogEntry
	subscribers []chan LogEntry
	mu          sync.Mutex
	logrus      *logrus.Logger
	logFile     *os.File
	logFilePath string
}

func GetInstance() *Logger {
	once.Do(func() {

		logrusLogger := logrus.New()

		logrusLogger.SetFormatter(&logrus.TextFormatter{
			FullTimestamp:   true,
			TimestampFormat: "2006-01-02 15:04:05",
		})

		logDir := getLogDirectory()
		if err := os.MkdirAll(logDir, 0755); err != nil {
			fmt.Printf("Failed to create log directory: %v\n", err)
		}

		logFilePath := filepath.Join(logDir, fmt.Sprintf("daily-pnl-%s.log", time.Now().Format("2006-01-02")))
		logFile, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			fmt.Printf("Failed to open log file: %v\n", err)
		} else {
			logrusLogger.SetOutput(logFile)
		}

		instance = &Logger{
			entries:     make([]LogEntry, 0),
			subscribers: make([]chan LogEntry, 0),
			logrus:      logrusLogger,
			logFile:     logFile,
			logFilePath: logFilePath,
		}

		instance.Debug("logging to %s", instance.GetLogFilePath())
	})
	return instance
}

func getLogDirectory() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "logs"
	}
	return filepath.Join(homeDir, ".daily-pnl", "logs")
}

func (l *Logger) GetLogFilePath() string {
	return l.logFilePath
}

func (l *Logger) Subscribe() chan LogEntry {
	l.mu.Lock()
	defer l.mu.Unlock()

	ch := make(chan LogEntry, 100)
	l.subscribers = append(l.subscribers, ch)

	go func() {
		for _, entry := range l.entries {
			ch <- entry
		}
	}()

	return ch
}

func (l *Logger) Unsubscribe(ch chan LogEntry) {
	l.mu.Lock()
	defer l.mu.Unlock()

	for i, subscriber := range l.subscribers {
		if subscriber == ch {
			l.subscribers = append(l.subscribers[:i], l.subscribers[i+1:]...)
			close(ch)
			break
		}
	}
}

func (l *Logger) GetAllLogs() []LogEntry {
	l.mu.Lock()
	defer l.mu.Unlock()

	result := make([]LogEntry, len(l.entries))
	copy(result, l.entries)
	return result
}

func (l *Logger) log(level LogLevel, format string, args ...interface{}) {
	entry := LogEntry{
		Timestamp: time.Now(),
		Level:     level,
		Message:   fmt.Sprintf(format, args...),
	}

	l.mu.Lock()
	l.entries = append(l.entries, entry)
	subscribers := make([]chan LogEntry, len(l.subscribers))
	copy(subscribers, l.subscribers)
	l.mu.Unlock()

	for _, subscriber := range subscribers {
		select {
		case subscriber <- entry:

		default:

			l.logrus.Warn("Subscriber channel is full, dropping log message")
		}
	}

	switch level {
	case Debug:
		l.logrus.Debugf(format, args...)
	case Info:
		l.logrus.Infof(format, args...)
	case Warning:
		l.logrus.Warnf(format, args...)
	case Error:
		l.logrus.Errorf(format, args...)
	}
}

func (l *Logger) Debug(format string, args ...interface{}) {
	l.log(Debug, format, args...)
}

func (l *Logger) Info(format string, args ...interface{}) {
	l.log(Info, format, args...)
}

func (l *Logger) Warning(format string, args ...interface{}) {
	l.log(Warning, format, args...)
}

func (l *Logger) Error(format string, args ...interface{}) {
	l.log(Error, format, args...)
}

func (l *Logger) SetLevel(level LogLevel) {
	switch level {
	case Debug:
		l.logrus.SetLevel(logrus.DebugLevel)
	case Info:
		l.logrus.SetLevel(logrus.InfoLevel)
	case Warning:
		l.logrus.SetLevel(logrus.WarnLevel)
	case Error:
		l.logrus.SetLevel(logrus.ErrorLevel)
	default:
		l.logrus.SetLevel(logrus.InfoLevel)
	}
}

func (l *Logger) Close() {
	if l.logFile != nil {
		l.logFile.Close()
	}
}
