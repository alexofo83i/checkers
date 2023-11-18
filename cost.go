package main

import "log"

//type Cost struct{}

func (playState *PlayState) Cost() int {
	if playState.cost != 0 {
		return playState.cost
	}
	cost := 0
	cntW := 0
	cntB := 0
	for ch := range playState.c2f {
		var whoisch byte = ch[0] // playState.c2f[ch][0] //
		if whoisch == black {
			cntB++
		} else {
			cntW++
		}
	}
	// need to increase priority of killing and defending from killing
	if playState.whodo == black {
		cost = 10*(cntB-cntW) + 3*(12-cntW)
	} else {
		cost = 10*(cntW-cntB) + 3*(12-cntB)
	}
	playState.cntB = cntB
	playState.cntW = cntW
	playState.cost = cost
	return playState.cost
}

func findIfKickStatesExistsBeforeOfBestState(playStateInit *PlayState) *PlayState {
	if playStateInit.nextStates == nil {
		return nil
	}

	nextStatesCnt := len(playStateInit.nextStates)
	if nextStatesCnt == 0 {
		return nil
	}

	var playStateKill *PlayState
	// check if we need to make a kick, if kick then no any way, so just kick
	playStateInit.Cost()

	var bestAlienScores int
	if convertWhoDo2WhoDoNext(playStateInit.whodo) == black {
		bestAlienScores = playStateInit.cntW
	} else {
		bestAlienScores = playStateInit.cntB
	}
	for i := nextStatesCnt - 1; i >= 0; i-- {
		var alienScoresAfterKick int
		if playStateInit.nextStates[i].whodo == black {
			alienScoresAfterKick = playStateInit.nextStates[i].cntW
		} else {
			alienScoresAfterKick = playStateInit.nextStates[i].cntB
		}
		if alienScoresAfterKick < bestAlienScores {
			playStateKill = playStateInit.nextStates[i]
			bestAlienScores = alienScoresAfterKick
		}
	}
	return playStateKill
}

// func findEndStates(playStateInit *PlayState) []*PlayState {
// 	endStates := make([]*PlayState, 0, 1000)
// 	// visitedStates is needed to reduce cases when same state is present in the different levels and so lead to loop
// 	visitedStates := make(map[uint32]*PlayState, 1000)

// 	nextStatesFinds := make([]*PlayState, 0, 10000)
// 	nextStatesFinds = append(nextStatesFinds, playStateInit)
// 	cnt := len(nextStatesFinds)
// 	for i := 0; i < cnt; i++ {
// 		nextstate := nextStatesFinds[i]
// 		_, visited := visitedStates[nextstate.Hashcode()]
// 		if !visited {
// 			// store nextstate as visited
// 			visitedStates[nextstate.Hashcode()] = nextstate
// 			// get count of next states
// 			cntNext := len(nextstate.nextStates)
// 			// if next state exist then continue deep dive into tree by next level
// 			if cntNext != 0 {
// 				// proceed with next states loop
// 				cnt += cntNext
// 				nextStatesFinds = append(nextStatesFinds, nextstate.nextStates...)
// 			} else {
// 				// mark state as end state for getting cost
// 				if nextstate.Cost() > 0 {
// 					endStates = append(endStates, nextstate)
// 				}
// 			}
// 		}
// 	}
// 	return endStates
// }

func findBestOfEndStates(playStateInit *PlayState, endStates []*PlayState) *PlayState {
	// find best end state
	var costBest int = 0
	playStateBest := endStates[0]
	// for i := endStatesCnt - 1; i >= 0; i-- {
	for i := range endStates {
		log.Default().Println("endstate: ", endStates[i].ToString())
		if endStates[i] != nil {

			cost := endStates[i].Cost()
			if endStates[i].whodo == playStateInit.whodo && cost < costBest || endStates[i].whodo != playStateInit.whodo && cost > costBest {
				playStateBest = endStates[i]
				costBest = cost
			}
		}
	}
	// back propagation from end state to init state
	log.Default().Println("backprop: ", playStateBest.ToString())
	for {
		playStateParent := playStateBest.prevState
		if playStateParent == nil {
			log.Fatal("Could not find parent state due to wrong caching. Please validate implementation of HashCode because len(visitedStates) > len( playStore.playStates) ")
		} else if playStateParent.Hashcode() != playStateInit.Hashcode() {
			playStateBest = playStateParent
			log.Default().Println("backprop: ", playStateBest.ToString())
		} else {
			break
		}
	}
	return playStateBest
}

// func findBestOfTheBestPlayState(playStateInit *PlayState) *PlayState {
// 	playStateInit.Cost()
// 	// check if we need to make a kick, if kick then no any way, so just kick
// 	playStateKill := findIfKickStatesExistsBeforeOfBestState(playStateInit)
// 	if playStateKill != nil {
// 		return playStateKill
// 	}
// 	// if no checkers were kicked then try find the best step

// 	endStates := findEndStates(playStateInit)
// 	if endStates == nil || len(endStates) == 0 {
// 		return nil
// 	}

// 	playStateBest := findBestOfEndStates(playStateInit, endStates)
// 	if playStateBest == nil {
// 		playStateBest = playStateInit.nextStates[0]
// 	}

// 	return playStateBest
// }

func getParentState(playState *PlayState) *PlayState {
	return playState.prevState
}
