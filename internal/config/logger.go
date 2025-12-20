package config

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/sirupsen/logrus"
)

var (
	Log     *logrus.Logger
	FileLog *logrus.Logger
	logFile *os.File
)

func InitLogger() {
	// Log for console
	Log = logrus.New()
	Log.SetReportCaller(true)
	Log.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
		CallerPrettyfier: func(f *runtime.Frame) (string, string) {
			fileName := filepath.Base(f.File)
			return "", fmt.Sprintf("%s:%d", fileName, f.Line)
		},
	})

	Log.SetOutput(os.Stderr)

	// Log for file
	filePath, err := GetLogFilePath()
	if err != nil {
		Log.Errorf("failed to get log file path: %v", err)
		return
	}

	FileLog = logrus.New()
	logFile, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		Log.Errorf("failed to get log file path: %v", err)
		return
	}

	FileLog.SetFormatter(&logrus.JSONFormatter{
		TimestampFormat: "2006-01-02 15:04:05",
	})
	FileLog.SetOutput(logFile)
}

func CloseLogger() {
	if logFile != nil {
		logFile.Close()
	}
}

func GetLogFilePath() (string, error) {
	exePath, err := os.Executable()
	if err != nil {
		return "", err
	}

	// The log file name must be followed by the current year
	year := time.Now().Year()
	logFileName := fmt.Sprintf("vote_%d.log", year)

	exeDir := filepath.Dir(exePath)
	logPath := filepath.Join(exeDir, "..", "..", "internal/log/", logFileName)
	return logPath, nil
}

func GetLogFilePathByYear(year string) (string, error) {
	exePath, err := os.Executable()
	if err != nil {
		return "", err
	}

	exeDir := filepath.Dir(exePath)
	logPath := filepath.Join(exeDir, "..", "..", fmt.Sprintf("log/vote_%s.log", year))

	return logPath, nil
}
