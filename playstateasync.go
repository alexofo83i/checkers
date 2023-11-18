package main

import (
	"log"
	"sync"
	"time"

	"golang.org/x/exp/maps"
)

// current implementation could work in parallel but couldn't exclude identical states
// [deepLevel 1 , len(workItemQueue) =  0]
// [deepLevel 2 , len(workItemQueue) =  7]
// [deepLevel 3 , len(workItemQueue) =  63]
// [deepLevel 4 , len(workItemQueue) =  528]
// [deepLevel 5 , len(workItemQueue) =  4683]
// [deepLevel 6 , len(workItemQueue) =  40111]
// [deepLevel 7 , len(workItemQueue) =  370135]

const (
	MAX_WORKERS             = 100
	MAX_CONSUMER_QUEUE_SIZE = 1000000
	MAX_PRODUCER_QUEUE_SIZE = 1000000
	MAX_FINAL_QUEUE_SIZE    = MAX_WORKERS + 1
	MAX_TIME_TO_WAIT        = 10 * time.Second
	MAX_LEVEL               = 4
)

type WorkItem struct {
	workPlayState *PlayState
	deepLevel     int
}

func sendProducedPlayStates(deepLevel int, nextStates []*PlayState, chanWorkItemsConsumerQueue chan *WorkItem, chanWorkItemsProducerQueue chan *WorkItem) {
	for i := range nextStates {
		workItemNextLevel := WorkItem{
			workPlayState: nextStates[i],
			deepLevel:     deepLevel + 1,
		}
		if deepLevel > MAX_LEVEL {
			chanWorkItemsConsumerQueue <- &workItemNextLevel
		} else {
			chanWorkItemsProducerQueue <- &workItemNextLevel
		}
	}
}

func getNextLevelStep(playStateInit *PlayState) *PlayState {

	chanWorkItemsProducerQueue := make(chan *WorkItem, MAX_PRODUCER_QUEUE_SIZE)
	chanWorkItemsConsumerQueue := make(chan *WorkItem, MAX_CONSUMER_QUEUE_SIZE)
	chanFinalPlayStatesQueue := make(chan *PlayState, MAX_FINAL_QUEUE_SIZE)
	var wgFinish sync.WaitGroup
	var wgStart sync.WaitGroup

	sizeMaxProducerQueue := len(chanWorkItemsProducerQueue)
	sizeMaxConsumerQueue := len(chanWorkItemsConsumerQueue)
	var sizeCurrProducerQueue, sizeCurrConsumerQueue int
	wgStart.Add(MAX_WORKERS)
	for i := 1; i <= MAX_WORKERS; i++ {
		wgFinish.Add(1)
		// i := i
		go func() {
			// threadId := i
			wgStart.Wait()
			var costBest int = 0
			var playStateBest *PlayState
			timerProducers := time.NewTimer(MAX_TIME_TO_WAIT)

			for {
				select {
				case workItem := <-chanWorkItemsProducerQueue:
					sizeCurrProducerQueue = len(chanWorkItemsProducerQueue)
					if sizeCurrProducerQueue > sizeMaxProducerQueue {
						sizeMaxProducerQueue = sizeCurrProducerQueue
					}
					// first of all we need check if it is present in cache
					// playStateFromCache, isCached := playStore.Get(workItem.workPlayState.Hashcode())
					// if isCached && playStateFromCache.nextStates != nil {
					// 	sendProducedPlayStates(workItem.deepLevel, playStateFromCache.nextStates, chanWorkItemsConsumerQueue, chanWorkItemsProducerQueue)
					// } else {
					// playState is not cached, so let's calculate possible states for each checker
					playStateFromQueue := workItem.workPlayState
					checkers := playStateFromQueue.getCheckersWhoDo(convertWhoDo2WhoDoNext(playStateFromQueue.whodo))
					allPossiblePlayStates := make([]*PlayState, 0, 16)
					for i := range checkers {
						somePossiblePlayStates := getPossiblePlayStates(playStateFromQueue, checkers[i], false, playStateFromQueue.level+1, nil)
						if len(somePossiblePlayStates) != 0 {
							allPossiblePlayStates = append(allPossiblePlayStates, somePossiblePlayStates...)
						}
					}
					playStateFromQueue.nextStates = allPossiblePlayStates
					// if we calculated next states then we need to store initial playstate linked with all next states
					playStore.Store(playStateFromQueue)

					sendProducedPlayStates(workItem.deepLevel, playStateFromQueue.nextStates, chanWorkItemsConsumerQueue, chanWorkItemsProducerQueue)
					// }
				case workItem := <-chanWorkItemsConsumerQueue:
					sizeCurrConsumerQueue = len(chanWorkItemsConsumerQueue)
					if sizeCurrConsumerQueue > sizeMaxConsumerQueue {
						sizeMaxConsumerQueue = sizeCurrConsumerQueue
					}
					// no needed to wait until all endstates will be calculated
					// we could find best state even right now!
					if playStateBest != nil {
						cost := workItem.workPlayState.Cost()
						if workItem.workPlayState.whodo == playStateInit.whodo && cost < costBest || workItem.workPlayState.whodo != playStateInit.whodo && cost > costBest {
							playStateBest = workItem.workPlayState
							costBest = cost
						}
					} else {
						playStateBest = workItem.workPlayState
						costBest = workItem.workPlayState.Cost()
					}
				case <-timerProducers.C:
					// no needs to calculate all states,
					// we could get only cached or piece of set of calculated states for current moment.
					// if no any final states calculated for required period of time
					// then random ( of first state ) should be taken because no back propagation is possible in this case
					// don't let your brain exploded
					// don't let your host too ( it could be overloaded )

					// add to channel of final states for performing a back propagation
					// within cost calculation to be able to choose correct state as next step
					if playStateBest != nil {
						chanFinalPlayStatesQueue <- playStateBest
					}
					wgFinish.Done()
					// log.Default().Println("threadId: ", threadId, ", sizeMaxConsumerQueue: ", sizeMaxConsumerQueue)
					// log.Default().Println("threadId: ", threadId, ", sizeMaxProducerQueue: ", sizeMaxProducerQueue)
					return
				}
			}
		}()
		wgStart.Done()
	}

	wgStart.Wait()
	// init first level producers
	chanWorkItemsProducerQueue <- &WorkItem{
		workPlayState: playStateInit,
		// initial deepLevel is 1,
		// do not miss with playState.level because we could have playState.level == 100
		// and we need next level that always will be +deepLevel which would be increased each time
		deepLevel: 1,
	}

	wgFinish.Wait()

	// close channel for final states queue to be able to go through the loop in range
	close(chanWorkItemsProducerQueue)
	close(chanWorkItemsConsumerQueue)
	close(chanFinalPlayStatesQueue)

	if playStateInit.nextStates == nil {
		log.Default().Println("playStateInit.nextStates == nil")
		return nil
	}
	if len(playStateInit.nextStates) == 0 {
		log.Default().Println("len(playStateInit.nextStates) == 0")
		return nil
	}

	if len(chanFinalPlayStatesQueue) == 0 {
		log.Default().Println("len(chanFinalPlayStatesQueue) == 0 ")
		return nil
	}

	playStateBest := findIfKickStatesExistsBeforeOfBestState(playStateInit)
	if playStateBest == nil {
		log.Default().Println("findIfKickStatesExistsBeforeOfBestState returned nil")
		// final states were collected for back propagation of Cost
		endStatesMap := make(map[uint32]*PlayState, len(chanFinalPlayStatesQueue))
		for i := 0; i <= len(chanFinalPlayStatesQueue); i++ {
			endState := <-chanFinalPlayStatesQueue
			if endState != nil {
				// check if exists or not. if not than adds.
				_, isExist := endStatesMap[endState.Hashcode()]
				if !isExist {
					endStatesMap[endState.Hashcode()] = endState
					// log.Default().Println("endstate[", i, "]: "+endState.ToString())
				}
			}
		}
		// at now we have reduced size of end states by excluding duplicates
		endStatesSlice := maps.Values(endStatesMap)
		playStateBest = findBestOfEndStates(playStateInit, endStatesSlice)
		if playStateBest == nil {
			playStateBest = playStateInit.nextStates[0]
			log.Default().Println("ohh!! findBestOfEndStates returned nil, so getting playStateInit.nextStates[0]", playStateBest.ToString())
		} else {
			log.Default().Println("yeaahooo!! findBestOfEndStates returned ", playStateBest.ToString())
		}
	}

	return playStateBest
}
