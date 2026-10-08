package user

import (
	"regexp"
	"slices"
	"strings"
	"sync"
)

// Service remembers a colour for each session and the sessions that are connected.
type Service interface {
	ByID(id string) (User, bool)
	Join(id string) (User, bool)
	Leave(id string)
	SetColour(id, colour string) (User, bool)
	Colours() []string
}

type service struct {
	mu     sync.Mutex
	known  map[string]user
	online map[string]struct{}
}

func NewService() Service {
	return &service{
		known:  make(map[string]user),
		online: make(map[string]struct{}),
	}
}

var sessionID = regexp.MustCompile(`^[A-Za-z0-9-]{8,64}$`)
var hexColour = regexp.MustCompile(`^#[0-9A-Fa-f]{6}$`)

func validID(id string) bool {
	return sessionID.MatchString(id)
}

// prepare stores a colour the first time an id is seen. The caller holds mu.
func (s *service) prepare(id string) (user, bool) {
	if !validID(id) {
		return user{}, false
	}
	if person, ok := s.known[id]; ok {
		return person, true
	}
	used := make(map[string]struct{}, len(s.known))
	for _, person := range s.known {
		used[person.colour] = struct{}{}
	}
	colour := ""
	for _, candidate := range palette {
		if _, taken := used[candidate]; !taken {
			colour = candidate
			break
		}
	}
	if colour == "" {
		colour = colourFor(id)
	}
	person := user{id: id, colour: colour}
	s.known[id] = person
	return person, true
}

func (s *service) ByID(id string) (User, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	person, ok := s.prepare(id)
	if !ok {
		return nil, false
	}
	return person, true
}

func (s *service) Join(id string) (User, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	person, ok := s.prepare(id)
	if !ok {
		return nil, false
	}
	s.online[id] = struct{}{}
	return person, true
}

func (s *service) Leave(id string) {
	s.mu.Lock()
	delete(s.online, id)
	s.mu.Unlock()
}

// SetColour stores a chosen #RRGGBB on a session the service already knows.
func (s *service) SetColour(id, colour string) (User, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	person, ok := s.known[id]
	if !ok || !hexColour.MatchString(colour) {
		return nil, false
	}
	person.colour = strings.ToUpper(colour)
	s.known[id] = person
	return person, true
}

func (s *service) Colours() []string {
	s.mu.Lock()
	defer s.mu.Unlock()
	colours := make([]string, 0, len(s.online))
	for id := range s.online {
		colours = append(colours, s.known[id].colour)
	}
	slices.Sort(colours)
	return colours
}
