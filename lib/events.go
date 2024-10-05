package lib

import (
	"fmt"
	"strings"
)

func FormatEvent(eventName string, data string) string {
	return fmt.Sprintf("event: %s\ndata: %s\n\n", eventName, data)
}

func CombineEvents(events ...string) string {
	var sb strings.Builder
	for _, message := range events {
		sb.WriteString(message)
	}

	return sb.String()
}
