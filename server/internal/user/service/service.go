package service

import "game_of_life/server/internal/user"

// Service looks up a user from an IP address and keeps the current set.
type Service interface {
	ByIP(ip string) user.User
	Colour(ip string) string
	Join(ip string) user.User
	Leave(ip string)
	Colours() []string
}

type service struct {
	users user.Users
}

func New() Service {
	return service{users: user.NewUsers()}
}

func (service) ByIP(ip string) user.User { return user.New(ip) }

func (s service) Colour(ip string) string { return s.ByIP(ip).Colour() }

func (s service) Join(ip string) user.User { return s.users.Join(ip) }

func (s service) Leave(ip string) { s.users.Leave(ip) }

func (s service) Colours() []string { return s.users.Colours() }
