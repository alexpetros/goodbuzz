package logger

import (
	"fmt"
	"log"
)

type logLevel int

const (
	DEBUG logLevel = 0
	INFO           = 1
)

var log_level logLevel = INFO

func InitLogger(dev_mode bool) {
	if dev_mode {
		log.SetFlags(0)
		log_level = DEBUG
	}
}

func Debug(msg string, args ...any) {
	if log_level <= DEBUG {
		message := fmt.Sprintf(msg, args...)
		log.Printf("[DEBUG] %s", message)
	}
}

func Info(msg string, args ...any) {
	message := fmt.Sprintf(msg, args...)
	log.Printf("[INFO] %s", message)
}

func Warn(msg string, args ...any) {
	message := fmt.Sprintf(msg, args...)
	log.Printf("[WARN] %s", message)
}

func Error(msg string, args ...any) {
	message := fmt.Sprintf(msg, args...)
	log.Printf("[ERROR] %s", message)
}

func Fatal(msg string, args ...any) {
	message := fmt.Sprintf(msg, args...)
	log.Fatalf("[FATAL] %s\n", message)
}
