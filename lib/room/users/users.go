package users

func SplitIntoTeams(players []*Player) ([]*Player, []*Player) {
	team1 := make([]*Player, 0)
	team2 := make([]*Player, 0)

	for _, player := range players {
		if player.Team == 1 {
			team1 = append(team1, player)
		} else {
			team2 = append(team2, player)
		}
	}

	return team1, team2
}
