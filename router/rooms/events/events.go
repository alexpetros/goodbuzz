package events

import (
	"fmt"
	"github.com/a-h/templ"
	"goodbuzz/lib"
	"strings"
)

func PlayerBuzzerEvent(buzzer templ.Component) string {
	data := lib.ToString(buzzer)
	return lib.FormatEventString("buzzer", data)
}

func PlayerLogEvent(message string) string {
	data := fmt.Sprintf("<div>%s</div>", message)
	return lib.FormatEventString("log", data)
}

func ModeratorStatusEvent(message string) string {
	data := fmt.Sprintf("<span>%s<span>", message)
	return lib.FormatEventString("status", data)
}

func ModeratorPlayerListEvent(players []string) string {
	var sb strings.Builder

	for _, name := range players {
		li := fmt.Sprintf("<li>%s", name)
		sb.WriteString(li)
	}

	return lib.FormatEventString("players", sb.String())
}

func ModeratorLogEvent(message string) string {
	data := fmt.Sprintf("<div>%s<div>", message)
	return lib.FormatEventString("log", data)
}

func TokenEvent(token templ.Component) string {
	data := lib.ToString(token)
	return lib.FormatEventString("token", data)
}
