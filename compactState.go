package main

import (
	"checkers/tools"
)

// // CellType описывает содержимое клетки
// type CellType uint8

// const (
// 	Empty      CellType = 0b00 // 0: пусто
// 	WhitePiece CellType = 0b01 // 1: белая шашка
// 	BlackPiece CellType = 0b11 // 3: чёрная шашка (цвет+занятость)
// )

// // Board — доска 8x8. Индексация: board[y][x], где x,y ∈ [0..7]
// type Board [8][8]CellType

// // PackBoard упаковывает состояние 32 тёмных клеток в uint64.
// // Кодирование: 2 бита на клетку.
// //
// //	бит 0: занятость (1 — есть шашка)
// //	бит 1: цвет     (1 — чёрная, 0 — белая)
// //
// // Порядок обхода: по строкам y=0..7, внутри строки x=0..7,
// // берутся только тёмные клетки (где (x+y)%2 == 1).
// func PackBoard(b Board) uint64 {
// 	var packed uint64 = 0
// 	index := 0 // номер тёмной клетки (0..31)

// 	for y := 0; y < 8; y++ {
// 		for x := 0; x < 8; x++ {
// 			// пропускаем светлые клетки
// 			if (x+y)%2 == 0 {
// 				continue
// 			}

// 			var bits uint64
// 			switch b[y][x] {
// 			case WhitePiece:
// 				bits = 0b01 // занято=1, цвет=0
// 			case BlackPiece:
// 				bits = 0b11 // занято=1, цвет=1
// 			case Empty:
// 				bits = 0b00 // занято=0, цвет=0
// 			}

// 			// сдвигаем 2 бита на позицию index*2 и вставляем через |
// 			packed |= bits << (index * 2)
// 			index++
// 		}
// 	}

// 	return packed
// }

// // UnpackBoard — обратная операция: восстанавливает доску из uint64
// func UnpackBoard(packed uint64) Board {
// 	var b Board
// 	index := 0

// 	for y := 0; y < 8; y++ {
// 		for x := 0; x < 8; x++ {
// 			if (x+y)%2 == 0 {
// 				continue
// 			}

// 			bits := (packed >> (index * 2)) & 0b11
// 			switch bits {
// 			case 0b01:
// 				b[y][x] = WhitePiece
// 			case 0b11:
// 				b[y][x] = BlackPiece
// 			default:
// 				b[y][x] = Empty
// 			}
// 			index++
// 		}
// 	}

// 	return b
// }

type CompactState struct {
	whodo  uint8
	cost   int
	packed uint64
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

func (state *PlayState) packState() *CompactState {
	cState := CompactState{whodo: state.whodo, cost: state.cost}
	var packed uint64
	for k, chId := range DarkFields {
		index := k + 1
		ch, isChecker := state.f2c[chId]
		var bits uint64
		if isChecker {
			if ch[0] == 'b' {
				bits = 0b11
			} else {
				bits = 0b01
			}
		}
		// else bits остаётся 0b00
		packed |= bits << (index * 2)
	}
	cState.packed = packed
	return &cState
}

func (cState *CompactState) unpackState() *PlayState {
	state := &PlayState{
		f2c:   make(map[string]string),
		c2f:   make(map[string]string),
		whodo: cState.whodo,
		cost:  cState.cost,
	}

	w, b := 1, 1
	for k, chId := range DarkFields {
		index := k + 1
		bits := (cState.packed >> (index * 2)) & 0b11

		switch bits {
		case 0b00:
			// пусто — ничего не делаем
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
		default: // 0b10
			panic("corrupted packed state: bits 0b10")
		}
	}

	return state
}
