package decks

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

const (
	MaxSavedDecks = 500
	MaxNameLength = 60
)

type SavedDeck struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Deck      []int     `json:"deck"`
	StructNo  int       `json:"structNo"`
	CreatedAt time.Time `json:"createdAt"`
}

// Store is a shared, file-backed list of decks anyone can save to or read from.
type Store struct {
	mutex sync.Mutex
	path  string
	decks []SavedDeck
}

func NewStore(path string) *Store {
	s := &Store{path: path}
	s.load()
	return s
}

func (s *Store) load() {
	data, err := os.ReadFile(s.path)
	if err != nil {
		return // no file yet, start empty
	}
	var loaded []SavedDeck
	if err := json.Unmarshal(data, &loaded); err != nil {
		return // corrupt file, start empty rather than crash the server
	}
	s.decks = loaded
}

func (s *Store) persist() error {
	if dir := filepath.Dir(s.path); dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}
	data, err := json.MarshalIndent(s.decks, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.path, data, 0644)
}

// List returns every saved deck, oldest first.
func (s *Store) List() []SavedDeck {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	out := make([]SavedDeck, len(s.decks))
	copy(out, s.decks)
	return out
}

// Add validates and appends a new deck, persisting the full list to disk.
func (s *Store) Add(name string, deck []int, structNo int) (SavedDeck, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return SavedDeck{}, errors.New("name is required")
	}
	if len(name) > MaxNameLength {
		return SavedDeck{}, errors.New("name is too long")
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()

	if len(s.decks) >= MaxSavedDecks {
		return SavedDeck{}, errors.New("saved deck list is full")
	}

	entry := SavedDeck{
		ID:        uuid.NewString(),
		Name:      name,
		Deck:      append([]int(nil), deck...),
		StructNo:  structNo,
		CreatedAt: time.Now(),
	}

	s.decks = append(s.decks, entry)
	if err := s.persist(); err != nil {
		s.decks = s.decks[:len(s.decks)-1] // roll back, the write failed
		return SavedDeck{}, err
	}

	return entry, nil
}
