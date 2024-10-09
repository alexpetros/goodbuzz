package users

type Player struct {
	name      string
	token     string
	channel   chan string
	is_locked bool
}

func NewPlayer(name string, token string, channel chan string) *Player {
	return &Player{name, token, channel, false}
}

func (player *Player) Name() string {
	return player.name
}

func (player *Player) Token() string {
	return player.token
}

func (player *Player) IsLocked() bool {
	return player.is_locked
}

func (player *Player) StatusString() string {
	if player.is_locked {
		return "Locked"
	} else {
		return "-"
	}
}

func (player *Player) Channel() chan string {
	return player.channel
}

func (player *Player) SetName(name string) {
	player.name = name
}

func (player *Player) Lock() {
	player.is_locked = true
}

func (player *Player) Unlock() {
	player.is_locked = false
}
