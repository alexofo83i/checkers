package main

import (
	"sync"
)

type PlayStore struct {
	playStates map[uint64]*CompactState
	mx         sync.RWMutex
}

var playStore = new(PlayStore).init()

func (playstore *PlayStore) init() *PlayStore {
	playstore.playStates = make(map[uint64]*CompactState, 1000000)
	return playstore
}

func (playstore *PlayStore) Get(key uint64) (*PlayState, bool) {
	playstore.mx.RLock()
	defer playstore.mx.RUnlock()
	v, ok := playStore.playStates[key]
	// if no OK then load from Postgresql database
	if v != nil {
		return v.unpackState(), ok
	}
	return nil, ok
}

func (playstore *PlayStore) Store(playState *PlayState) {
	playstore.mx.Lock()
	defer playstore.mx.Unlock()
	playstore.playStates[playState.Hashcode()] = playState.packState()
}

// func hash(s string) uint32 {
// 	h := fnv.New32a()
// 	h.Write([]byte(s))
// 	return h.Sum32()
// }

func (playState *PlayState) Hashcode() uint64 {
	return playState.getCachedPacked()
}
