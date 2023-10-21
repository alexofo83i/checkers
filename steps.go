package main

// type StepsRequest struct {
// 	gameId  string
// 	checker string
// }

// func getStepsHandler(wr http.ResponseWriter, r *http.Request) {
// 	// decode input or return error
// 	var game *Game
// 	var stepsRequest *StepsRequest

// 	// body, _ := ioutil.ReadAll(r.Body)
// 	// err := json.Unmarshal(body, &gameState)
// 	err := json.NewDecoder(r.Body).Decode(&stepsRequest)
// 	if err != nil {
// 		log.Fatal("StepsRequest couldn't be parsed due to error: ", err)
// 	}
// 	game, found := gamesStore.GetGame(stepsRequest.gameId)
// 	if !found {
// 		log.Fatal("Gamestate was not found due to error: ", err)
// 	}
// 	game.
// }

type NextStepVisiter func(checker string, stepTo string, chKicked string, iskick bool, iskickbykick bool)

func fillPlayStateByNextSteps(playState *PlayState, table *Table, whodo byte) {
	checkers := playState.getCheckersWhoDo(whodo)
	ch2Step := make(map[string]NextStep, 10)
	for _, ch := range checkers {
		getPossiblePlayStates(playState, ch, false, playState.level+1,
			func(checker string, step string, chKicked string, iskick bool, iskickbykick bool) {
				ns, exist := ch2Step[checker]
				if !exist {
					ns = NextStep{Checker: checker}
					ns.NextSteps = make([]string, 0, 4)
					ns.NextKicks = make([]string, 0, 4)
				}
				if !iskick {
					ns.NextSteps = append(ns.NextSteps, step)
				} else {
					if !iskickbykick {
						ns.NextKicks = append(ns.NextKicks, chKicked+"-"+step+",")
					} else {
						lastKick := ns.NextKicks[len(ns.NextKicks)-1]
						lastKick = lastKick + chKicked + "-" + step + ","
						ns.NextKicks[len(ns.NextKicks)-1] = lastKick
					}
				}
				ch2Step[checker] = ns
			})
	}
	table.NextSteps = nil
	table.NextSteps = make([]NextStep, len(ch2Step))
	for _, st := range ch2Step {
		table.NextSteps = append(table.NextSteps, st)
	}
}

func (playState *PlayState) getCheckersWhoDo(whodo byte) []string {
	var checkers = make([]string, 0, len(playState.c2f))
	for key := range playState.c2f {
		if key[0] == whodo {
			checkers = append(checkers, key)
		}
	}
	return checkers
}

type PossibleStep struct {
	checker string
	step    string
	kick    string
}

func getPossiblePlayStates(playState *PlayState, checker string, kickonly bool, level int, nextStepVisiter NextStepVisiter) []*PlayState {
	playStateNewSlice := make([]*PlayState, 0, 10)
	fieldFromStr := playState.c2f[checker]
	fieldFrom := getFieldOnBoard(fieldFromStr)
	// for i := len(fieldFrom.fieldsAround) - 1; i >= 0; i-- {
	for i := range fieldFrom.fieldsAround {
		var ch = playState.f2c[fieldFrom.fieldsAround[i].Yx()]
		if !kickonly && ch == "" {
			// you can do it
			if canMakeStep(fieldFrom, fieldFrom.fieldsAround[i], checker[0]) {
				playStateNew := playState.MakeNextStep(checker, fieldFrom.fieldsAround[i], level)
				playStateNew.Cost()
				if nextStepVisiter != nil {
					nextStepVisiter(checker, fieldFrom.fieldsAround[i].Yx(), "", false, false)
				}
				playStateNewSlice = append(playStateNewSlice, playStateNew)
			}
		} else if ch != "" && ch[0] != checker[0] {
			// field is busy
			// need check by whom  if alien then kick
			fieldForKick := getFieldForKick(fieldFrom, fieldFrom.fieldsAround[i])
			if fieldForKick != nil {
				yxForKick := fieldForKick.Yx()
				chForKick := playState.f2c[yxForKick]
				if chForKick == "" {
					// if we make kick then we need check fields around fieldForKick and do it recursively
					playStateKick := playState.MakeNextKick(checker, fieldForKick, ch, level)
					if kickonly {
						playStateKick.whodo = playState.whodo
					}
					playStateKick.Cost()
					if nextStepVisiter != nil {
						nextStepVisiter(checker, yxForKick, ch, true, kickonly)
					}
					nextKicks := getPossiblePlayStates(playStateKick, checker, true, level, nextStepVisiter)
					if len(nextKicks) != 0 {
						playStateNewSlice = append(playStateNewSlice, nextKicks...)
					} else {
						playStateNewSlice = append(playStateNewSlice, playStateKick)
					}
				}
			}
		}
	}
	//log.Default().Println("ch # ", checker, " lead to ", len(playStateNewSlice), " possible states")
	return playStateNewSlice
}
