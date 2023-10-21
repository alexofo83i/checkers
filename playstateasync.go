package main

import (
	"sync"
	"time"

	"golang.org/x/exp/maps"
)

const (
	MAX_WORKERS          = 1
	MAX_PROD_QUEUE       = 100000
	MAX_END_QUEUE        = 10000
	MAX_TIME_TO_WAIT     = 10 * time.Second //time.Second
	MAX_LEVEL        int = 3
)

// 1a -> w12
// 1-8 a-h -> w/b 1-12
// 1 + 1 -> 1 + 2 = 5
// 1
var (
	workItemQueue = new(WorkPool).Init()
)

func getNextLevelStep(playStateInit *PlayState) *PlayState {

	chanWorkItemsProducerQueue := make(chan WorkItem, MAX_PROD_QUEUE)
	chanFinalPlayStatesQueue := make(chan *PlayState, MAX_END_QUEUE)
	var wgFinish sync.WaitGroup
	// var wgStart sync.WaitGroup

	// wgStart.Add(MAX_WORKERS)
	for i := 1; i <= MAX_WORKERS; i++ {
		wgFinish.Add(1)
		go func() {
			// wgStart.Wait()
			timerProducers := time.NewTimer(MAX_TIME_TO_WAIT)
			for {
				select {
				case workItem := <-chanWorkItemsProducerQueue:
					// check if we already calculated states on required level
					if workItem.deepLevel > MAX_LEVEL {
						// add to channel of final states for performing a back propagation
						//  within cost calculation to be able to choose correct state as next step
						chanFinalPlayStatesQueue <- workItem.workPlayState
					}
					//  else {
					// 	// ok, we have playstate that should be calculated or  get from cache
					// 	workItemQueue <- workItem
					// }
				case <-timerProducers.C:
					// no needs to calculate all states,
					// we could get only cached or piece of set of calculated states for current moment.
					// if no any final states calculated for required period of time
					// then random ( of first state ) should be taken because no back propagation is possible in this case
					// don't let your brain exploded
					// don't let your host too ( it could be overloaded )
					wgFinish.Done()
					return
				}
			}
		}()
		// wgStart.Done()
	}

	// init first level producers
	workItemQueue <- WorkItem{
		workPlayState: playStateInit,
		// initial deepLevel is 1,
		// do not miss with playState.level because we could have playState.level == 100
		// and we need next level that always will be +deepLevel which would be increased each time
		deepLevel:    1,
		callbackChan: chanWorkItemsProducerQueue,
	}

	wgFinish.Wait()
	// wgStart.Wait()
	close(chanWorkItemsProducerQueue)
	// close channel for final states queue to be able to go through the loop in range
	close(chanFinalPlayStatesQueue)

	if playStateInit.nextStates == nil || len(playStateInit.nextStates) == 0 {
		return nil
	}

	if len(chanFinalPlayStatesQueue) == 0 {
		return nil
	}

	// final states were collected for back propagation of Cost
	endStatesMap := make(map[uint32]*PlayState, len(chanFinalPlayStatesQueue))
	for i := 0; i <= len(chanFinalPlayStatesQueue); i++ {
		endState := <-chanFinalPlayStatesQueue
		_, isExist := endStatesMap[endState.Hashcode()]
		if !isExist {
			endStatesMap[endState.Hashcode()] = endState
			// log.Default().Println("endstate[", i, "]: "+endState.ToString())
		}

	}
	endStatesSlice := maps.Values(endStatesMap)
	playStateBest := findBestOfEndStates(playStateInit, endStatesSlice)
	if playStateBest == nil {
		playStateBest = playStateInit.nextStates[0]
	}
	return playStateBest
}
