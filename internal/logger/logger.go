package logger

import (
	"fmt"
	"log"
	"os"

	"github.com/LinharesAron/jotunn/internal/ui"
)

var fileLogger *log.Logger
var enableFile bool = false

func Init(path string) {
	if path == "" {
		return
	}

	logFile, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		Error("Failed to create log file: %v", err)
		return
	}

	fileLogger = log.New(logFile, "", log.LstdFlags)
	enableFile = true
}

func logWithColor(colorCode string, prefix string, msg string, fixed bool) {
	ui.GetUI().SendLogEvent(prefix, colorCode, msg, fixed)
}

func Info(format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	logWithColor("\033[33m", "", msg, false)
	if enableFile {
		fileLogger.Printf("[INFO] %s", msg)
	}
}

func Error(format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	logWithColor("\033[31m", "[ERROR] ", msg, false)
	if enableFile {
		fileLogger.Println("[ERROR]", fmt.Sprint(args...))
	}
}

func Success(format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	logWithColor("\033[32m", "[SUCCESS] ", msg, true)
	if enableFile {
		fileLogger.Println("[SUCCESS]", fmt.Sprint(args...))
	}
}

func Warn(format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	logWithColor("\033[35m", "[WARN] ", msg, false)
	if enableFile {
		fileLogger.Println("[WARN]", fmt.Sprint(args...))
	}
}
