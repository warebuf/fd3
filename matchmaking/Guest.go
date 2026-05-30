package matchmaking

import (
	"fmt"
	"sync"
	"time"
)

type Guest struct {
	Expiry       time.Time
	Deck         []int
	StartingHand []int
	StructNo     int
}

type GuestStore struct {
	mutex sync.Mutex
	store map[string]Guest
}

func NewGuestStore() *GuestStore {
	gs := &GuestStore{
		store: make(map[string]Guest),
	}

	// background cleaner
	go func() {
		for range time.Tick(time.Minute * 15) {
			gs.cleanup()
		}
	}()

	return gs
}

func (guestMap *GuestStore) Add(id string, expiry time.Time, deck []int, startingHand []int, structNo int) {
	guestMap.mutex.Lock()
	defer guestMap.mutex.Unlock()
	guestMap.store[id] = Guest{
		Expiry:       expiry,
		Deck:         deck,
		StartingHand: startingHand,
		StructNo:     structNo,
	}
}

func (guestMap *GuestStore) Remove(id string) {
	guestMap.mutex.Lock()
	defer guestMap.mutex.Unlock()
	delete(guestMap.store, id)
}

func (guestMap *GuestStore) Get(id string) (Guest, bool) {
	guestMap.mutex.Lock()
	defer guestMap.mutex.Unlock()
	guest, ok := guestMap.store[id]
	if !ok || time.Now().After(guest.Expiry) {
		delete(guestMap.store, id)
		return Guest{}, false
	}
	return guest, true
}

func (guestMap *GuestStore) cleanup() {
	guestMap.mutex.Lock()
	defer guestMap.mutex.Unlock()
	now := time.Now()
	for id, guest := range guestMap.store {
		if now.After(guest.Expiry) {
			delete(guestMap.store, id)
		}
	}
}

func (guestMap *GuestStore) Print() {
	fmt.Println("Printing guestMap")
	guestMap.mutex.Lock()
	defer guestMap.mutex.Unlock()
	for id, guest := range guestMap.store {
		fmt.Println("GuestStore", id, guest)
	}
}
