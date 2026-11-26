package main

import (
	"checkers/tools"
)

type CompactState struct {
	ID    int64
	Whodo uint8
	Cost  int
	Place uint32
	Color uint32
}

// There are 32 fields and only 2 bits per field in uint64.
var DarkFields = [32]string{
	// row 8
	"8b", "8d", "8f", "8h",
	// row 7
	"7a", "7c", "7e", "7g",
	// row 6
	"6b", "6d", "6f", "6h",
	// row 5
	"5a", "5c", "5e", "5g",
	// row 4
	"4b", "4d", "4f", "4h",
	// row 3
	"3a", "3c", "3e", "3g",
	// row 2
	"2b", "2d", "2f", "2h",
	// row 1
	"1a", "1c", "1e", "1g",
}

const signBit = uint64(1) << 63

// EncodeID: (place, color) -> int64 PK, биекция без коллизий
func EncodeID(place, color uint32) int64 {
	packed := (uint64(place) << 32) | uint64(color) // 0 .. 2^64-1
	return int64(packed - signBit)                  // -2^63 .. 2^63-1
}

// DecodeID: int64 PK -> (place, color)
func DecodeID(id int64) (place, color uint32) {
	packed := uint64(id) + signBit // 0 .. 2^64-1
	return uint32(packed >> 32), uint32(packed & 0xFFFFFFFF)
}

func (state *PlayState) packState() *CompactState {
	var place, color uint32
	for k, chId := range DarkFields {
		ch, isChecker := state.f2c[chId]
		if !isChecker {
			continue
		}
		var bits uint32
		if ch[0] == 'b' {
			bits = 0b11
		} else {
			bits = 0b01
		}
		if k < 16 {
			place |= bits << (k * 2)
		} else {
			color |= bits << ((k - 16) * 2)
		}
	}
	return &CompactState{
		ID:    EncodeID(place, color),
		Whodo: state.whodo,
		Cost:  state.cost,
		Place: place,
		Color: color,
	}
}

func (cState *CompactState) unpackState() *PlayState {
	state := &PlayState{
		f2c:   make(map[string]string),
		c2f:   make(map[string]string),
		whodo: cState.Whodo,
		cost:  cState.Cost,
	}
	w, b := 1, 1
	for k, chId := range DarkFields {
		var bits uint32
		if k < 16 {
			bits = (cState.Place >> (k * 2)) & 0b11
		} else {
			bits = (cState.Color >> ((k - 16) * 2)) & 0b11
		}
		switch bits {
		case 0b00:
			// пусто
		case 0b11:
			ch := "b" + tools.IntToString(b)
			b++
			state.f2c[chId] = ch
			state.c2f[ch] = chId
		case 0b01:
			ch := "w" + tools.IntToString(w)
			w++
			state.f2c[chId] = ch
			state.c2f[ch] = chId
		default:
			panic("corrupted packed state: bits 0b10")
		}
	}
	return state
}
