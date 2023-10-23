package main

import (
	"checkers/tools"
	"log"
	"strconv"
	"strings"
)

const (
	black byte = 'b'
	white byte = 'w'
)

// 1a -> w12
// 1-8 a-h -> w/b 1-12
// 1 + 1 -> 1 + 2 = 5
// 1

type PlayState struct {
	f2c        map[string]string // each field could be represented by 5 bits ( 32 states )
	c2f        map[string]string // each kind of checker could be represented by 1 bit ( w / b ) and 1 bit ( checker or queen )
	whodo      byte
	prevState  *PlayState
	nextStates []*PlayState
	cost       int // could be 0 or 1, so could be replaced by 1 byte
	cntW       int
	cntB       int
	hashcode   uint32 // need replace by uint32 ( 4 byte )
	level      int
	strCached  string
	history    strings.Builder
}

func (state *PlayState) ToString() string {
	if state.strCached != "" {
		return state.strCached
	}
	var b strings.Builder
	b.WriteString("level=")
	b.WriteString(strconv.Itoa(state.level))
	b.WriteString(", whodo=")
	b.WriteString(string(state.whodo))
	b.WriteString(", hashcode=")
	b.WriteString(tools.UInt32ToString(state.Hashcode()))
	b.WriteString(", cost=")
	b.WriteString(tools.IntToString(state.Cost()))
	b.WriteString(", history=")
	b.WriteString(state.history.String())
	b.WriteString("\n")
	for j := 'a'; j <= 'h'; j++ {
		b.WriteString("_")
		b.WriteString(string(j))
		b.WriteString("_")
	}
	b.WriteString("\n")
	for i := 8; i >= 1; i-- {
		k := 1
		// b.WriteString("|")
		for j := 'a'; j <= 'h'; j++ {
			if i%2+1 == k%2+1 || i%2 == k%2 {
				chId := tools.IntToString(i) + string(j)
				ch, isChecker := state.f2c[chId]
				if isChecker {
					b.WriteString(ch)
					if len(ch) < 3 {
						b.WriteString(" ")
					}
				} else {
					b.WriteString("   ")
				}
			} else {
				b.WriteString("░░░")
			}
			k++
		}
		b.WriteString("|")
		b.WriteString(tools.IntToString(i))
		b.WriteString("\n")
	}
	b.WriteString("\n")
	state.strCached = b.String()
	return state.strCached
}

func UInt32ToString(u uint32) {
	panic("unimplemented")
}

func playWithMe(game *Game) {
	if game == nil {
		return
	}

	playStateInit := game.State.convertGame2PlayState()
	playStateNext := getNextLevelStep(playStateInit)
	// playStateNext => Game
	var tableStateNew *Table
	if playStateNext == nil {
		log.Default().Println("sorry we could not calculate next step, so returned the same", playStateInit.ToString())
		playStateNext = playStateInit
	}
	tableStateNew = playStateNext.convertPlayState2Table(game.State.Name)
	// if playStateNext.whodo != black {
	// 	tableStateNew.Whodo = "checkers-black"
	// } else {
	// 	tableStateNew.Whodo = "checkers-white"
	// }
	fillPlayStateByNextSteps(playStateNext, tableStateNew, playStateInit.whodo)

	game.State = *tableStateNew
}

func NewPlayState(whodo byte, checkersCount int) *PlayState {
	playStateInit := PlayState{whodo: whodo}
	playStateInit.c2f = make(map[string]string, checkersCount)
	playStateInit.f2c = make(map[string]string, checkersCount)
	return &playStateInit
}

func (playState *PlayState) MakeNextStep(ch string, fieldTo *Field, level int) *PlayState {
	var playStateCopy *PlayState
	playStateCopy = &PlayState{whodo: convertWhoDo2WhoDoNext(playState.whodo)}
	playStateCopy.c2f = copyMap(playState.c2f)
	playStateCopy.f2c = copyMap(playState.f2c)
	playStateCopy.prevState = playState
	playStateCopy.level = level
	playStateCopy.nextStates = nil
	fieldSrc := playStateCopy.c2f[ch]
	fieldDst := fieldTo.Yx()
	if playStateCopy.history.Len() != 0 {
		playStateCopy.history.WriteString(",")
	}
	playStateCopy.history.WriteString(ch)
	playStateCopy.history.WriteString(":")
	playStateCopy.history.WriteString(fieldSrc)
	playStateCopy.history.WriteString("-")
	playStateCopy.history.WriteString(fieldDst)
	playStateCopy.c2f[ch] = fieldDst
	delete(playStateCopy.f2c, fieldSrc)
	playStateCopy.f2c[fieldDst] = ch
	return playStateCopy
}

func copyMap[K, V comparable](m map[K]V) map[K]V {
	result := make(map[K]V, len(m))
	for k, v := range m {
		result[k] = v
	}
	return result
}

func (playState *PlayState) MakeNextKick(ch string, fieldTo *Field, ch2kick string, level int) *PlayState {
	playStateKick := playState.MakeNextStep(ch, fieldTo, level)
	// playStateKick.kicked = playStateKick.kicked + 1
	// playStateKick.cost = playStateKick.cost + 1
	field2Kick := playStateKick.c2f[ch2kick]

	playStateKick.history.WriteString(":")
	playStateKick.history.WriteString(ch2kick)
	playStateKick.history.WriteString(":")
	playStateKick.history.WriteString(field2Kick)
	delete(playStateKick.f2c, field2Kick)
	delete(playStateKick.c2f, ch2kick)

	return playStateKick
}

// func getNextLevelStep(playStateInit *PlayState, level uint32) {
// 	if level > maxLevel {
// 		return
// 	}

// 	checkers := playStateInit.getCheckersWhoDo(convertWhoDo2WhoDoNext(playStateInit.whodo))

// 	c1 := make(chan []*PlayState, maxdop)
// 	cnt := 0
// 	for i := len(checkers) - 1; i >= 0; i-- {
// 		i := i
// 		go func() {
// 			c1 <- getPossiblePlayStates(playStateInit, checkers[i], false, level, nil)
// 		}()
// 		cnt++
// 	}
// 	allPossiblePlayStates := make([]*PlayState, 0, 100)
// 	for i := 0; i < cnt; i++ {
// 		somePossiblePlayStates := <-c1
// 		if len(somePossiblePlayStates) != 0 {
// 			allPossiblePlayStates = append(allPossiblePlayStates, somePossiblePlayStates...)
// 		}
// 	}
// 	playStateInit.nextStates = allPossiblePlayStates
// 	cnt = 0
// 	c2 := make(chan int, maxdop)
// 	for i := len(allPossiblePlayStates) - 1; i >= 0; i-- {
// 		// possiblePlayState := possiblePlayState
// 		possibleStateFromStore, cached := playStore.Get(allPossiblePlayStates[i].Hashcode())
// 		if !cached || possibleStateFromStore.nextStates == nil {
// 			i := i
// 			level := level
// 			go func() {
// 				getNextLevelStep(allPossiblePlayStates[i], level+1)
// 				c2 <- 1
// 			}()
// 			cnt++
// 		} else {
// 			allPossiblePlayStates[i] = possibleStateFromStore
// 		}
// 	}
// 	for i := 0; i < cnt; i++ {
// 		<-c2
// 	}
// 	playStore.Store(playStateInit)
// }
