package rooms

import (
	"fmt"
	"github.com/a-h/templ"
	"goodbuzz/lib"
	"strings"
)

func formatEvent(eventName string, data string) string {
	return fmt.Sprintf("event: %s\ndata: %s\n\n", eventName, data)
}

func CombineEvents(events ...string) string {
	var sb strings.Builder
	for _, message := range events {
		sb.WriteString(message)
	}

	return sb.String()
}

func PlayerBuzzerEvent(buzzer templ.Component) string {
	data := lib.ToString(buzzer)
	return formatEvent("buzzer", data)
}

func PlayerLogEvent(message string) string {
	data := fmt.Sprintf("<div>%s</div>", message)
	return formatEvent("log", data)
}

func ModeratorStatusEvent(message string) string {
	data := fmt.Sprintf("<span>%s<span>", message)
	return formatEvent("status", data)
}

func ModeratorPlayerListEvent(players []string) string {
	var sb strings.Builder

	for _, name := range players {
		li := fmt.Sprintf("<li>%s", name)
		sb.WriteString(li)
	}

	return formatEvent("players", sb.String())
}

func ModeratorLogEvent(message string) string {
	data := fmt.Sprintf("<div>%s<div>", message)
	return formatEvent("log", data)
}
