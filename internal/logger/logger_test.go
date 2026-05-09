package logger

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func setupTestLogger(t *testing.T) string {
	t.Helper()
	tmpDir := t.TempDir()

	// 重置包级别变量
	mu.Lock()
	if logFile != nil {
		logFile.Close()
		logFile = nil
	}
	logDir = filepath.Join(tmpDir, "logs")
	currentDate = ""
	logLevel = DEBUG
	mu.Unlock()

	os.MkdirAll(logDir, 0755)
	openLogFile()
	return logDir
}

func TestLoggerInit(t *testing.T) {
	tmpDir := t.TempDir()
	origUserProfile := os.Getenv("USERPROFILE")
	origHome := os.Getenv("HOME")
	os.Setenv("USERPROFILE", tmpDir)
	os.Setenv("HOME", tmpDir)
	defer func() {
		os.Setenv("USERPROFILE", origUserProfile)
		os.Setenv("HOME", origHome)
		if logFile != nil {
			logFile.Close()
			logFile = nil
		}
	}()

	os.Setenv("UUPT_LOG_LEVEL", "DEBUG")
	defer os.Unsetenv("UUPT_LOG_LEVEL")

	Init()

	expectedDir := filepath.Join(tmpDir, ".uupt-open-cli", "logs")
	if _, err := os.Stat(expectedDir); os.IsNotExist(err) {
		t.Errorf("日志目录未创建: %s", expectedDir)
	}
}

func TestLoggerWritesToFile(t *testing.T) {
	dir := setupTestLogger(t)
	defer func() {
		if logFile != nil {
			logFile.Close()
			logFile = nil
		}
	}()

	Info("test message")

	// 验证日志文件存在且包含内容
	dateStr := time.Now().Format("2006-01-02")
	logPath := filepath.Join(dir, "uupt-open-cli."+dateStr+".log")

	data, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("读取日志文件失败: %v", err)
	}

	content := string(data)
	if !strings.Contains(content, "test message") {
		t.Errorf("日志文件不包含预期内容: %s", content)
	}
}

func TestLogLevel(t *testing.T) {
	dir := setupTestLogger(t)

	// 设置为 ERROR 级别
	mu.Lock()
	logLevel = ERROR
	mu.Unlock()

	defer func() {
		if logFile != nil {
			logFile.Close()
			logFile = nil
		}
	}()

	Debug("debug msg")
	Info("info msg")
	Warn("warn msg")
	Error("error msg")

	dateStr := time.Now().Format("2006-01-02")
	logPath := filepath.Join(dir, "uupt-open-cli."+dateStr+".log")

	data, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("读取日志文件失败: %v", err)
	}

	content := string(data)
	if strings.Contains(content, "debug msg") {
		t.Error("ERROR级别不应记录debug日志")
	}
	if strings.Contains(content, "info msg") {
		t.Error("ERROR级别不应记录info日志")
	}
	if strings.Contains(content, "warn msg") {
		t.Error("ERROR级别不应记录warn日志")
	}
	if !strings.Contains(content, "error msg") {
		t.Error("ERROR级别应记录error日志")
	}
}

func TestLogFormat(t *testing.T) {
	dir := setupTestLogger(t)
	defer func() {
		if logFile != nil {
			logFile.Close()
			logFile = nil
		}
	}()

	Info("format test")

	dateStr := time.Now().Format("2006-01-02")
	logPath := filepath.Join(dir, "uupt-open-cli."+dateStr+".log")

	data, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("读取日志文件失败: %v", err)
	}

	content := string(data)
	// 验证格式: [YYYY-MM-DD HH:MM:SS] [LEVEL] message
	if !strings.Contains(content, "["+dateStr) {
		t.Errorf("日志应包含日期: %s", content)
	}
	if !strings.Contains(content, "[INFO]") {
		t.Errorf("日志应包含级别标识: %s", content)
	}
	if !strings.Contains(content, "format test") {
		t.Errorf("日志应包含消息内容: %s", content)
	}
}

func TestLogf(t *testing.T) {
	dir := setupTestLogger(t)
	defer func() {
		if logFile != nil {
			logFile.Close()
			logFile = nil
		}
	}()

	Infof("hello %s %d", "world", 42)

	dateStr := time.Now().Format("2006-01-02")
	logPath := filepath.Join(dir, "uupt-open-cli."+dateStr+".log")

	data, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("读取日志文件失败: %v", err)
	}

	content := string(data)
	if !strings.Contains(content, "hello world 42") {
		t.Errorf("Infof格式化不正确: %s", content)
	}
}
