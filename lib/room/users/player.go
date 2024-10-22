package users

type Player struct {
	Name     string
	Token    string
	IsLocked bool
}

func NewPlayer(name string, token string, isLocked bool) *Player {
	return &Player{name, token, isLocked}
}

func (player *Player) StatusString() string {
	if player.IsLocked {
		return "Locked"
	} else {
		return "-"
	}
}

func (player *Player) SetName(name string) {
	player.Name = name
}

func (player *Player) Lock() {
	player.IsLocked = true
}

func (player *Player) Unlock() {
	player.IsLocked = false
}
