package main

// import (
// 	"log"
// )

var (
	MAX_CHAN_SIZE       int = 100000
	MAX_THREADS_IN_POOL int = 1
)

type WorkPool struct {
}

type WorkItem struct {
	workPlayState *PlayState
	deepLevel     int
	callbackChan  chan WorkItem
}

func (*WorkPool) Init() chan WorkItem {
	workItemQueue := make(chan WorkItem, MAX_CHAN_SIZE)
	for i := 1; i <= MAX_THREADS_IN_POOL; i++ {
		go calculateNextLevelStates(i, workItemQueue)
	}
	return workItemQueue
}

func debuglog(workThreadId int, msg ...any) {
	//log.Default().Println("tid #", workThreadId, ":", msg)
}

func calculateNextLevelStates(workThreadId int, workItemQueue chan WorkItem) {
	debuglog(workThreadId, "started")

	for {
		workItem := <-workItemQueue
		debuglog(workThreadId, "workItem = ", workItem.workPlayState.ToString(), " <-workItemQueue")
		// first of all we need check if it is present in cache
		// playStateInit, isCached := playStore.Get(workItem.workPlayState.Hashcode())
		// debuglog(workThreadId, "isCached = ", isCached)
		// if isCached && playStateInit.nextStates != nil {
		// 	for i := range playStateInit.nextStates {
		// 		debuglog(workThreadId, "send Next State from Cached")
		// 		workItem.callbackChan <- WorkItem{
		// 			workPlayState: playStateInit.nextStates[i],
		// 			deepLevel:     workItem.deepLevel + 1,
		// 			callbackChan:  workItem.callbackChan,
		// 		}
		// 	}
		// } else {
		debuglog(workThreadId, " begin calculate")
		// playState is not cached, so let's calculate possible states for each checker
		playStateInit := workItem.workPlayState
		checkers := playStateInit.getCheckersWhoDo(convertWhoDo2WhoDoNext(playStateInit.whodo))
		allPossiblePlayStates := make([]*PlayState, 0, 16)
		for i := range checkers {
			somePossiblePlayStates := getPossiblePlayStates(playStateInit, checkers[i], false, playStateInit.level+1, nil)
			if len(somePossiblePlayStates) != 0 {
				allPossiblePlayStates = append(allPossiblePlayStates, somePossiblePlayStates...)
			}
		}
		playStateInit.nextStates = allPossiblePlayStates
		debuglog(workThreadId, "finish calculate")
		// if we calculated next states then we need to store initial playstate linked with all next states
		// playStore.Store(playStateInit)

		// for i := len(allPossiblePlayStates) - 1; i >= 0; i-- {
		// for i, _ := range allPossiblePlayStates {
		for i := range allPossiblePlayStates {

			workItemNextLevel := WorkItem{
				workPlayState: allPossiblePlayStates[i],
				// deepLevel is increased because of next level from current work
				deepLevel:    workItem.deepLevel + 1,
				callbackChan: workItem.callbackChan,
			}
			if workItem.deepLevel > MAX_LEVEL {
				workItem.callbackChan <- workItemNextLevel
			} else {
				workItemQueue <- workItemNextLevel
			}
		}
		// }
	}
}
