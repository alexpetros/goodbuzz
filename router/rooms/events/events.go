package events

import (
	"fmt"
	"goodbuzz/lib"
	"goodbuzz/router/rooms/users"
	"strings"
)

func LoadingBuzzerEvent() string {
	return lib.FormatEventComponent("buzzer", LoadingBuzzer())
}

func WaitingBuzzerEvent() string {
	return lib.FormatEventComponent("buzzer", WaitingBuzzer())
}

func LockedBuzzerEvent() string {
	return lib.FormatEventComponent("buzzer", LockedBuzzer())
}

func ReadyBuzzerEvent() string {
	return lib.FormatEventComponent("buzzer", ReadyBuzzer())
}

func PlayerLogEvent(message string) string {
	data := fmt.Sprintf("<div>%s</div>", message)
	return lib.FormatEventString("log", data)
}

func ModeratorStatusEvent(message string) string {
	data := fmt.Sprintf("<span>%s<span>", message)
	return lib.FormatEventString("status", data)
}

func PlayerListEvent(players []string) string {
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

func ModeratorPlayerControlsEvent(players []*users.Player) string {
	component := ModeratorPlayerControls(players)
	return lib.FormatEventComponent("controls", component)
}

func TokenEvent(token string) string {
	data := lib.ToString(TokenInput(token))
	return lib.FormatEventString("token", data)
}
