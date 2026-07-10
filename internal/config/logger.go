package config

import (
	"io"
	"log"
	"os"
	"path/filepath"
	"sync"
)

var (
	logger     *log.Logger
	logFile    *os.File
	logLevel   string
	logMu      sync.RWMutex
)

const (
	LevelDebug = "DEBUG"
	LevelInfo  = "INFO"
	LevelWarn  = "WARN"
	LevelError = "ERROR"
)

func InitLogger() error {
	logMu.Lock()
	defer logMu.Unlock()

	logLevel = Cfg.Log.Level

	if err := os.MkdirAll(Cfg.Log.Dir, 0755); err != nil {
		return err
	}

	logPath := filepath.Join(Cfg.Log.Dir, "caddy-webui.log")
	f, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return err
	}
	logFile = f

	multiWriter := io.MultiWriter(os.Stdout, f)
	logger = log.New(multiWriter, "", log.LstdFlags|log.Lshortfile)

	return nil
}

func SetLogLevel(level string) {
	logMu.Lock()
	defer logMu.Unlock()
	logLevel = level
}

func shouldLog(level string) bool {
	logMu.RLock()
	defer logMu.RUnlock()

	levels := map[string]int{
		LevelDebug: 0,
		LevelInfo:  1,
		LevelWarn:  2,
		LevelError: 3,
	}

	current, ok1 := levels[logLevel]
	req, ok2 := levels[level]
	if !ok1 || !ok2 {
		return true
	}
	return req >= current
}

func Debug(format string, v ...interface{}) {
	if shouldLog(LevelDebug) && logger != nil {
		logger.Printf("[DEBUG] "+format, v...)
	}
}

func Info(format string, v ...interface{}) {
	if shouldLog(LevelInfo) && logger != nil {
		logger.Printf("[INFO] "+format, v...)
	}
}

func Warn(format string, v ...interface{}) {
	if shouldLog(LevelWarn) && logger != nil {
		logger.Printf("[WARN] "+format, v...)
	}
}

func Error(format string, v ...interface{}) {
	if shouldLog(LevelError) && logger != nil {
		logger.Printf("[ERROR] "+format, v...)
	}
}

func CloseLogger() {
	if logFile != nil {
		logFile.Close()
	}
}
