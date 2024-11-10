package users

type Player struct {
	Name     string
	Team int64
	Token    string
	IsLocked bool
}

func NewPlayer(name string, team int64, token string, isLocked bool) *Player {
	return &Player{name, team, token, isLocked}
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

func (player *Player) SetTeam(team int64) {
	player.Team = team
}

func (player *Player) Lock() {
	player.IsLocked = true
}

func (player *Player) Unlock() {
	player.IsLocked = false
}
