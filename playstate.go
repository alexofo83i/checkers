package main

import (
	"checkers/tools"
	"log"
	"strconv"
	"strings"
)

type PlayState struct {
	f2c        map[string]string
	c2f        map[string]string
	whodo      byte
	prevState  *PlayState
	nextStates []*PlayState
	cost       int
	cntW       int
	cntB       int
	hashcode   uint32
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

func playWithMe(game *Game) {
	if game == nil {
		return
	}

	playStateInit := game.State.convertGame2PlayState()
	playStateNext := getNextLevelStep(playStateInit)
	var tableStateNew *Table
	if playStateNext == nil {
		log.Default().Println("sorry we could not calculate next step, so returned the same", playStateInit.ToString())
		playStateNext = playStateInit
	}
	tableStateNew = playStateNext.convertPlayState2Table(game.State.Name)
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
	field2Kick := playStateKick.c2f[ch2kick]

	playStateKick.history.WriteString(":")
	playStateKick.history.WriteString(ch2kick)
	playStateKick.history.WriteString(":")
	playStateKick.history.WriteString(field2Kick)
	delete(playStateKick.f2c, field2Kick)
	delete(playStateKick.c2f, ch2kick)

	return playStateKick
}
