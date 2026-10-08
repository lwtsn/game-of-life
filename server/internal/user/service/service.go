package service

import "game_of_life/server/internal/user"

// Service looks up a user from an IP address.
type Service interface {
	ByIP(ip string) user.User
	Colour(ip string) string
}

type service struct{}

func New() Service { return service{} }

func (service) ByIP(ip string) user.User { return user.New(ip) }

func (s service) Colour(ip string) string { return s.ByIP(ip).Colour() }
