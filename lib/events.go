package lib

import (
	"fmt"
	"github.com/a-h/templ"
	"strings"
)

func FormatEventString(eventName string, data string) string {
	return fmt.Sprintf("event: %s\ndata: %s\n\n", eventName, data)
}

func FormatEventComponent(eventName string, component templ.Component) string {
	data := ToString(component)
	return fmt.Sprintf("event: %s\ndata: %s\n\n", eventName, data)
}

func CombineEvents(events ...string) string {
	var sb strings.Builder
	for _, message := range events {
		sb.WriteString(message)
	}

	return sb.String()
}
