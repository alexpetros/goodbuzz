package users

type Player struct {
	name    string
	channel chan string
	is_locked bool
}

func NewPlayer(name string, channel chan string) *Player {
	return &Player{name, channel, false}
}

func (player Player) Name() string {
	return player.name
}

func (player *Player) SetName(name string) {
	player.name = name
}

func (player Player) Channel() chan string {
	return player.channel
}

type Moderator struct {
	channel chan string
}

func NewModerator(channel chan string) *Moderator {
	return &Moderator{channel}
}

func (moderator Moderator) Channel() chan string {
	return moderator.channel
}
