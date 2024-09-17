package logger

import (
	"fmt"
	"log"
)

func Debug (msg string, args ...any) {
	message := fmt.Sprintf(msg, args...)
	log.Printf("[DEBUG] %s", message)
}

func Info (msg string, args ...any) {
	message := fmt.Sprintf(msg, args...)
	log.Printf("[INFO] %s", message)
}

func Warn (msg string, args ...any) {
	message := fmt.Sprintf(msg, args...)
	log.Printf("[WARN] %s", message)
}

func Error (msg string, args ...any) {
	message := fmt.Sprintf(msg, args...)
	log.Printf("[ERROR] %s", message)
}

func Fatal (msg string, args ...any) {
	message := fmt.Sprintf(msg, args...)
	log.Fatalf("[FATAL] %s\n", message)
}
