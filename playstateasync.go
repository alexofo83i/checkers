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
	MAX_TIME_TO_WAIT        = 30 * time.Second
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
		go func() {
			wgStart.Wait()
			var costBest int = 0
			var playStateBest *PlayState
			timerProducers := time.NewTimer(MAX_TIME_TO_WAIT)

			// Целевой игрок – тот, чей ход наступает после хода текущего (playStateInit.whodo)
			targetWhodo := convertWhoDo2WhoDoNext(playStateInit.whodo)

			for {
				select {
				case workItem := <-chanWorkItemsProducerQueue:
					sizeCurrProducerQueue = len(chanWorkItemsProducerQueue)
					if sizeCurrProducerQueue > sizeMaxProducerQueue {
						sizeMaxProducerQueue = sizeCurrProducerQueue
					}
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
					playStore.Store(playStateFromQueue)

					sendProducedPlayStates(workItem.deepLevel, playStateFromQueue.nextStates, chanWorkItemsConsumerQueue, chanWorkItemsProducerQueue)

				case workItem := <-chanWorkItemsConsumerQueue:
					sizeCurrConsumerQueue = len(chanWorkItemsConsumerQueue)
					if sizeCurrConsumerQueue > sizeMaxConsumerQueue {
						sizeMaxConsumerQueue = sizeCurrConsumerQueue
					}
					// Рассматриваем только состояния после хода текущего игрока (ход перешёл к противнику)
					if workItem.workPlayState.whodo == targetWhodo {
						cost := workItem.workPlayState.Cost()
						if playStateBest == nil || cost < costBest {
							playStateBest = workItem.workPlayState
							costBest = cost
						}
					}

				case <-timerProducers.C:
					if playStateBest != nil {
						chanFinalPlayStatesQueue <- playStateBest
					}
					wgFinish.Done()
					return
				}
			}
		}()
		wgStart.Done()
	}

	wgStart.Wait()
	chanWorkItemsProducerQueue <- &WorkItem{
		workPlayState: playStateInit,
		deepLevel:     1,
	}

	wgFinish.Wait()

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
		endStatesMap := make(map[uint32]*PlayState, len(chanFinalPlayStatesQueue))
		for i := 0; i <= len(chanFinalPlayStatesQueue); i++ {
			endState := <-chanFinalPlayStatesQueue
			if endState != nil {
				_, isExist := endStatesMap[endState.Hashcode()]
				if !isExist {
					endStatesMap[endState.Hashcode()] = endState
				}
			}
		}
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
