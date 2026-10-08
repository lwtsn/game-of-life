package user

import (
	"slices"
	"sync"
)

// Users is the people currently here, one per IP address.
type Users interface {
	Join(host string) User
	Leave(host string)
	Colours() []string
}

type users struct {
	mu      sync.Mutex
	current map[string]User
}

func NewUsers() Users {
	return &users{current: make(map[string]User)}
}

func (u *users) Join(host string) User {
	person := New(host)
	u.mu.Lock()
	u.current[person.IP()] = person
	u.mu.Unlock()
	return person
}

func (u *users) Leave(host string) {
	person := New(host)
	u.mu.Lock()
	delete(u.current, person.IP())
	u.mu.Unlock()
}

func (u *users) Colours() []string {
	u.mu.Lock()
	defer u.mu.Unlock()
	colours := make([]string, 0, len(u.current))
	for _, person := range u.current {
		colours = append(colours, person.Colour())
	}
	slices.Sort(colours)
	return colours
}
