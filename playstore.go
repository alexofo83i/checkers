package main

import (
	"hash/fnv"
	"sort"
	"sync"
)

type PlayStore struct {
	playStates map[uint32]*PlayState
	mx         sync.RWMutex
}

var playStore = new(PlayStore).init()

func (playstore *PlayStore) init() *PlayStore {
	playstore.playStates = make(map[uint32]*PlayState, 1000000)
	return playstore
}

func (playstore *PlayStore) Get(key uint32) (*PlayState, bool) {
	playstore.mx.RLock()
	defer playstore.mx.RUnlock()
	v, ok := playStore.playStates[key]
	return v, ok
	// return nil, false
}

func (playstore *PlayStore) Store(playState *PlayState) {
	playstore.mx.Lock()
	defer playstore.mx.Unlock()
	playstore.playStates[playState.Hashcode()] = playState
}

func hash(s string) uint32 {
	h := fnv.New32a()
	h.Write([]byte(s))
	return h.Sum32()
}

func (playState *PlayState) Hashcode() uint32 {
	if playState.hashcode != 0 {
		return playState.hashcode
	}
	val := string(playState.whodo)
	//+ string(playState.level)
	i := 0
	keys := make([]string, len(playState.f2c), len(playState.f2c))
	for k := range playState.f2c {
		keys[i] = k
		i++
	}
	sort.Strings(keys)
	for i := range keys {
		val += keys[i] + playState.f2c[keys[i]]
	}
	playState.hashcode = hash(val)
	return playState.hashcode
}
