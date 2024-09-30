package rooms

import (
	"fmt"
	"github.com/a-h/templ"
	"goodbuzz/lib"
)

func formatEvent(eventName string, data string) string {
	return fmt.Sprintf("event: %s\ndata: %s\n\n", eventName, data)
}

func BuzzerEvent(buzzer templ.Component) string {
	data := lib.ToString(buzzer)
	return formatEvent("buzzer", data)
}

func PlayerLogEvent(message string) string {
	data := fmt.Sprintf("<div>%s</div>", message)
	return formatEvent("log", data)
}

func ModeratorLogEvent(message string) string {
	data := fmt.Sprintf("<span>%s<span>", message)
	return formatEvent("status", data)
}
