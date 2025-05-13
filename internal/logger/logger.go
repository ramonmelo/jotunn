package logger

import (
	"fmt"
	"log"
	"os"
	"sync"

	"github.com/LinharesAron/jotunn/internal/ui"
)

var fileLogger *log.Logger
var enableFile bool = false
var logMu sync.Mutex

func Init(path string) {
	if path == "" {
		return
	}

	logFile, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		fmt.Fprintf(os.Stdout, "\033[31mFailed to create log file: %v\033[0m\n", err)
		return
	}

	fileLogger = log.New(logFile, "", log.LstdFlags)
	enableFile = true
}

func logWithColor(colorCode string, prefix string, msg string) {
	logMu.Lock()
	defer logMu.Unlock()

	ui.UI.LogLine(prefix, colorCode, msg)
	if Progress != nil {
		Progress.renderInline()
	}
}

func Info(format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	logWithColor("\033[33m", "", msg)
	if enableFile {
		fileLogger.Printf("[INFO] %s", msg)
	}
}

func Error(format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	logWithColor("\033[31m", "[ERROR] ", msg)
	if enableFile {
		fileLogger.Println("[ERROR]", fmt.Sprint(args...))
	}
}

func Success(format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	logWithColor("\033[32m", "[SUCCESS] ", msg)
	if enableFile {
		fileLogger.Println("[SUCCESS]", fmt.Sprint(args...))
	}
}

func Warn(format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	logWithColor("\033[35m", "[WARN] ", msg)
	if enableFile {
		fileLogger.Println("[WARN]", fmt.Sprint(args...))
	}
}
