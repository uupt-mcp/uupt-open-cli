package logger

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type Level int

const (
	DEBUG Level = iota
	INFO
	WARN
	ERROR
)

var (
	mu          sync.Mutex
	logFile     *os.File
	currentDate string
	logLevel    Level
	logDir      string
)

func Init() {
	logDir = filepath.Join(getHomeDir(), "logs")
	os.MkdirAll(logDir, 0755)

	// Set log level from environment
	levelStr := strings.ToUpper(os.Getenv("UUPT_LOG_LEVEL"))
	switch levelStr {
	case "DEBUG":
		logLevel = DEBUG
	case "WARN":
		logLevel = WARN
	case "ERROR":
		logLevel = ERROR
	default:
		logLevel = INFO
	}

	// Clean old log files (older than 30 days)
	cleanOldLogs()

	// Open today's log file
	openLogFile()
}

func getHomeDir() string {
	var home string
	if os.Getenv("USERPROFILE") != "" {
		home = os.Getenv("USERPROFILE")
	} else {
		home = os.Getenv("HOME")
	}
	return filepath.Join(home, ".uupt-open-cli")
}

func cleanOldLogs() {
	entries, err := os.ReadDir(logDir)
	if err != nil {
		return
	}
	cutoff := time.Now().AddDate(0, 0, -30)
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			continue
		}
		if info.ModTime().Before(cutoff) {
			os.Remove(filepath.Join(logDir, entry.Name()))
		}
	}
}

func openLogFile() {
	now := time.Now()
	dateStr := now.Format("2006-01-02")
	currentDate = dateStr
	filename := fmt.Sprintf("uupt-open-cli.%s.log", dateStr)
	path := filepath.Join(logDir, filename)

	var err error
	logFile, err = os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		// Silently fail - logs only go to file
		logFile = nil
	}
}

func checkRotation() {
	now := time.Now()
	dateStr := now.Format("2006-01-02")
	if dateStr != currentDate {
		if logFile != nil {
			logFile.Close()
		}
		openLogFile()
	}
}

func writeLog(level Level, levelStr string, msg string) {
	if level < logLevel {
		return
	}

	mu.Lock()
	defer mu.Unlock()

	checkRotation()

	if logFile == nil {
		return
	}

	now := time.Now()
	line := fmt.Sprintf("[%s] [%s] %s\n", now.Format("2006-01-02 15:04:05"), levelStr, msg)
	logFile.WriteString(line)
}

func Debug(msg string) { writeLog(DEBUG, "DEBUG", msg) }
func Info(msg string)  { writeLog(INFO, "INFO", msg) }
func Warn(msg string)  { writeLog(WARN, "WARN", msg) }
func Error(msg string) { writeLog(ERROR, "ERROR", msg) }

func Debugf(format string, args ...interface{}) { Debug(fmt.Sprintf(format, args...)) }
func Infof(format string, args ...interface{})  { Info(fmt.Sprintf(format, args...)) }
func Warnf(format string, args ...interface{})  { Warn(fmt.Sprintf(format, args...)) }
func Errorf(format string, args ...interface{}) { Error(fmt.Sprintf(format, args...)) }
