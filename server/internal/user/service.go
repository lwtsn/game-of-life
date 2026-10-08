package user

import (
	"slices"
	"sync"
)

// Service looks up a user from an IP address and keeps the current set.
type Service interface {
	ByIP(ip string) User
	Colour(ip string) string
	Join(ip string) User
	Leave(ip string)
	Colours() []string
}

type service struct {
	mu      sync.Mutex
	current map[string]User
}

func NewService() Service {
	return &service{current: make(map[string]User)}
}

func (*service) ByIP(ip string) User { return New(ip) }

func (s *service) Colour(ip string) string { return s.ByIP(ip).Colour() }

func (s *service) Join(host string) User {
	person := New(host)
	s.mu.Lock()
	s.current[person.IP()] = person
	s.mu.Unlock()
	return person
}

func (s *service) Leave(host string) {
	person := New(host)
	s.mu.Lock()
	delete(s.current, person.IP())
	s.mu.Unlock()
}

func (s *service) Colours() []string {
	s.mu.Lock()
	defer s.mu.Unlock()
	colours := make([]string, 0, len(s.current))
	for _, person := range s.current {
		colours = append(colours, person.Colour())
	}
	slices.Sort(colours)
	return colours
}
