package users

type Moderator struct {
	channel chan string
}

func NewModerator(channel chan string) *Moderator {
	return &Moderator{channel}
}

func (moderator Moderator) Channel() chan string {
	return moderator.channel
}
