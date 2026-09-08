package main

import "log"

const (
	black byte = 'b'
	white byte = 'w'

	WEIGHT_MATERIAL      = 100
	WEIGHT_POTENTIAL     = 5
	WEIGHT_THREAT        = 15
	WEIGHT_DEFENSE       = 20
	WEIGHT_ATTACK        = 12
	WEIGHT_VULNERABILITY = -30
	WEIGHT_CENTER        = 3
	WEIGHT_MOBILITY      = 2
	WEIGHT_EDGE          = -2
	WEIGHT_BACK_ROW      = 4
	WEIGHT_EXCHANGE      = 25
	WEIGHT_PIN           = 15
)

func (playState *PlayState) Cost() int {
	if playState.cost != 0 {
		return playState.cost
	}

	if playState.cntW == 0 && playState.cntB == 0 {
		playState.countPieces()
	}

	materialDiff := playState.calculateMaterialDiff()
	threats := playState.calculateThreats()
	vulnerabilities := playState.calculateVulnerabilities()
	defense := playState.calculateDefense()
	attackPotential := playState.calculateAttackPotential()
	centerDiff := playState.calculateCenterControl()
	mobilityDiff := playState.calculateMobility()
	edgePenalty := playState.calculateEdgePenalty()
	backRowBonus := playState.calculateBackRowBonus()
	pinnedPenalty := playState.calculatePinnedPenalty()

	var cost int
	if playState.whodo == black {
		cost = WEIGHT_MATERIAL*materialDiff +
			WEIGHT_POTENTIAL*(12-playState.cntW) +
			WEIGHT_THREAT*threats.black +
			WEIGHT_VULNERABILITY*vulnerabilities.black +
			WEIGHT_DEFENSE*defense.black +
			WEIGHT_ATTACK*attackPotential.black +
			WEIGHT_CENTER*centerDiff +
			WEIGHT_MOBILITY*mobilityDiff +
			WEIGHT_EDGE*edgePenalty +
			WEIGHT_BACK_ROW*backRowBonus +
			WEIGHT_PIN*pinnedPenalty.black
	} else {
		cost = WEIGHT_MATERIAL*(-materialDiff) +
			WEIGHT_POTENTIAL*(12-playState.cntB) +
			WEIGHT_THREAT*threats.white +
			WEIGHT_VULNERABILITY*vulnerabilities.white +
			WEIGHT_DEFENSE*defense.white +
			WEIGHT_ATTACK*attackPotential.white +
			WEIGHT_CENTER*(-centerDiff) +
			WEIGHT_MOBILITY*(-mobilityDiff) +
			WEIGHT_EDGE*(-edgePenalty) +
			WEIGHT_BACK_ROW*(-backRowBonus) +
			WEIGHT_PIN*pinnedPenalty.white
	}

	playState.cost = cost
	return playState.cost
}

func (playState *PlayState) calculateMaterialDiff() int {
	baseDiff := playState.cntW - playState.cntB
	positionalValueW := 0
	positionalValueB := 0

	for pos, ch := range playState.c2f {
		posValue := playState.getPieceValue(pos)
		if ch[0] == white {
			positionalValueW += posValue
		} else {
			positionalValueB += posValue
		}
	}
	_ = baseDiff
	return (playState.cntB - playState.cntW) + (positionalValueB - positionalValueW)
}

func (playState *PlayState) getPieceValue(pos string) int {
	ch, exists := playState.c2f[pos]
	if !exists {
		return 0
	}
	color := ch[0]

	value := 1

	if isCenterSquare(pos) {
		value += 2
	}
	if len(pos) > 0 && (pos[0] == 'a' || pos[0] == 'h') {
		value -= 1
	}
	if playState.canAttackFrom(pos) {
		value += 2
	}

	row := pos[1]
	if color == white {
		switch row {
		case '7':
			value += 3
		case '6':
			value += 2
		case '5':
			value += 1
		}
		if row == '1' && (pos[0] == 'b' || pos[0] == 'g') {
			value += 1
		}
	} else {
		switch row {
		case '2':
			value += 3
		case '3':
			value += 2
		case '4':
			value += 1
		}
		if row == '8' && (pos[0] == 'b' || pos[0] == 'g') {
			value += 1
		}
	}
	return value
}

func (playState *PlayState) calculateThreats() struct{ white, black int } {
	threats := struct{ white, black int }{0, 0}
	for pos, ch := range playState.c2f {
		if ch[0] == white {
			threats.white += playState.countAttacksFrom(pos)
		} else {
			threats.black += playState.countAttacksFrom(pos)
		}
	}
	return threats
}

func (playState *PlayState) countAttacksFrom(pos string) int {
	attacks := 0
	color := playState.c2f[pos][0]
	opponent := white
	if color == white {
		opponent = black
	}
	directions := []struct{ dx, dy int }{
		{-1, 1}, {1, 1}, {-1, -1}, {1, -1},
	}
	for _, dir := range directions {
		victimPos := shiftPos(pos, dir.dx, dir.dy)
		if victimPos == "" {
			continue
		}
		jumpPos := shiftPos(victimPos, dir.dx, dir.dy)
		if jumpPos == "" {
			continue
		}
		if victim, exists := playState.c2f[victimPos]; exists && victim[0] == opponent {
			if _, exists := playState.c2f[jumpPos]; !exists {
				attacks++
			}
		}
	}
	return attacks
}

func (playState *PlayState) calculateVulnerabilities() struct{ white, black int } {
	vuln := struct{ white, black int }{0, 0}
	for pos, ch := range playState.c2f {
		if ch[0] == white {
			vuln.white += playState.countVulnerabilityAt(pos)
		} else {
			vuln.black += playState.countVulnerabilityAt(pos)
		}
	}
	return vuln
}

func (playState *PlayState) countVulnerabilityAt(pos string) int {
	color := playState.c2f[pos][0]
	opponent := white
	if color == white {
		opponent = black
	}
	vulnerabilities := 0
	directions := []struct{ dx, dy int }{
		{-1, 1}, {1, 1}, {-1, -1}, {1, -1},
	}
	for _, dir := range directions {
		attackerPos := shiftPos(pos, -dir.dx, -dir.dy)
		if attackerPos == "" {
			continue
		}
		if attacker, exists := playState.c2f[attackerPos]; exists && attacker[0] == opponent {
			if playState.canAttackFrom(attackerPos) {
				vulnerabilities++
			}
		}
	}
	return vulnerabilities
}

func (playState *PlayState) calculateDefense() struct{ white, black int } {
	defense := struct{ white, black int }{0, 0}
	for pos, ch := range playState.c2f {
		if ch[0] == white {
			defense.white += playState.countDefenders(pos)
		} else {
			defense.black += playState.countDefenders(pos)
		}
	}
	return defense
}

func (playState *PlayState) countDefenders(pos string) int {
	color := playState.c2f[pos][0]
	defenders := 0
	directions := []struct{ dx, dy int }{
		{-1, -1}, {1, -1}, {-1, 1}, {1, 1},
	}
	for _, dir := range directions {
		defenderPos := shiftPos(pos, -dir.dx, -dir.dy)
		if defenderPos == "" {
			continue
		}
		if defender, exists := playState.c2f[defenderPos]; exists && defender[0] == color {
			defenders++
		}
	}
	return defenders
}

func (playState *PlayState) calculateAttackPotential() struct{ white, black int } {
	potential := struct{ white, black int }{0, 0}
	for pos, ch := range playState.c2f {
		if ch[0] == white && playState.canAttackFrom(pos) {
			potential.white++
		} else if ch[0] == black && playState.canAttackFrom(pos) {
			potential.black++
		}
	}
	return potential
}

func (playState *PlayState) canAttackFrom(pos string) bool {
	return playState.countAttacksFrom(pos) > 0
}

func (playState *PlayState) calculatePinnedPenalty() struct{ white, black int } {
	pinned := struct{ white, black int }{0, 0}
	for pos, ch := range playState.c2f {
		if !playState.hasAnyMove(pos) {
			if ch[0] == white {
				pinned.white++
			} else {
				pinned.black++
			}
		}
	}
	return pinned
}

func (playState *PlayState) hasAnyMove(pos string) bool {
	ch, exists := playState.c2f[pos]
	if !exists {
		return false
	}
	color := ch[0]
	var directions []struct{ dx, dy int }
	if color == white {
		directions = []struct{ dx, dy int }{{-1, 1}, {1, 1}}
	} else {
		directions = []struct{ dx, dy int }{{-1, -1}, {1, -1}}
	}
	if playState.canAttackFrom(pos) {
		return true
	}
	for _, dir := range directions {
		newPos := shiftPos(pos, dir.dx, dir.dy)
		if newPos != "" {
			if _, exists := playState.c2f[newPos]; !exists {
				return true
			}
		}
	}
	return false
}

func shiftPos(pos string, dx, dy int) string {
	if len(pos) != 2 {
		return ""
	}
	col := pos[0]
	row := pos[1]
	newCol := byte(int(col) + dx)
	newRow := byte(int(row) + dy)
	if newCol < 'a' || newCol > 'h' || newRow < '1' || newRow > '8' {
		return ""
	}
	return string([]byte{newCol, newRow})
}

func isCenterSquare(pos string) bool {
	if len(pos) != 2 {
		return false
	}
	col := pos[0]
	row := pos[1]
	centerCols := map[byte]bool{'c': true, 'd': true, 'e': true, 'f': true}
	centerRows := map[byte]bool{'3': true, '4': true, '5': true, '6': true}
	return centerCols[col] && centerRows[row]
}

func (playState *PlayState) countPieces() {
	cntW := 0
	cntB := 0
	for _, ch := range playState.c2f {
		if ch[0] == black {
			cntB++
		} else {
			cntW++
		}
	}
	playState.cntW = cntW
	playState.cntB = cntB
}

func (playState *PlayState) calculateCenterControl() int {
	centerW := 0
	centerB := 0
	centerSquares := map[string]bool{
		"c3": true, "d3": true, "e3": true, "f3": true,
		"c4": true, "d4": true, "e4": true, "f4": true,
		"c5": true, "d5": true, "e5": true, "f5": true,
		"c6": true, "d6": true, "e6": true, "f6": true,
	}
	for pos, ch := range playState.c2f {
		if !centerSquares[pos] {
			continue
		}
		if ch[0] == white {
			centerW++
		} else {
			centerB++
		}
	}
	return centerW - centerB
}

func (playState *PlayState) calculateMobility() int {
	mobilityW := 0
	mobilityB := 0
	for pos, ch := range playState.c2f {
		if ch[0] == white {
			mobilityW += playState.countSimpleMoves(pos)
		} else {
			mobilityB += playState.countSimpleMoves(pos)
		}
	}
	return mobilityW - mobilityB
}

func (playState *PlayState) countSimpleMoves(pos string) int {
	color := playState.c2f[pos][0]
	moves := 0
	var directions []struct{ dx, dy int }
	if color == white {
		directions = []struct{ dx, dy int }{{-1, 1}, {1, 1}}
	} else {
		directions = []struct{ dx, dy int }{{-1, -1}, {1, -1}}
	}
	for _, dir := range directions {
		newPos := shiftPos(pos, dir.dx, dir.dy)
		if newPos != "" {
			if _, exists := playState.c2f[newPos]; !exists {
				moves++
			}
		}
	}
	return moves
}

func (playState *PlayState) calculateEdgePenalty() int {
	edgeW := 0
	edgeB := 0
	edgeCols := map[byte]bool{'a': true, 'h': true}
	for pos, ch := range playState.c2f {
		if len(pos) < 2 {
			continue
		}
		col := pos[0]
		if !edgeCols[col] {
			continue
		}
		if ch[0] == white {
			edgeW++
		} else {
			edgeB++
		}
	}
	return edgeW - edgeB
}

func (playState *PlayState) calculateBackRowBonus() int {
	backRowW := 0
	backRowB := 0
	for pos, ch := range playState.c2f {
		if len(pos) < 2 {
			continue
		}
		row := pos[1]
		if ch[0] == white && row == '1' {
			backRowW++
		} else if ch[0] == black && row == '8' {
			backRowB++
		}
	}
	return backRowW - backRowB
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
	if playStateKill != nil {
		logStateJSON(playStateKill, "winning_kick_move")
	}
	return playStateKill
}

func findBestOfEndStates(playStateInit *PlayState, endStates []*PlayState) *PlayState {
	if len(endStates) == 0 {
		return nil
	}

	targetWhodo := convertWhoDo2WhoDoNext(playStateInit.whodo)

	var playStateBest *PlayState
	var costBest int
	initialized := false

	for _, state := range endStates {
		if state == nil || state.whodo != targetWhodo {
			continue
		}
		cost := state.Cost()
		if !initialized || cost < costBest {
			playStateBest = state
			costBest = cost
			initialized = true
			// Логируем лучшее конечное состояние
			logStateJSON(playStateBest, "best_end_state")
		}
	}

	if playStateBest == nil {
		playStateBest = endStates[0]
		costBest = playStateBest.Cost()
	}

	// Обратная пропагация
	for {
		parent := playStateBest.prevState
		if parent == nil {
			log.Fatal("Could not find parent state")
		} else if parent.Hashcode() != playStateInit.Hashcode() {
			playStateBest = parent
			logStateJSON(playStateBest, "backprop")
		} else {
			break
		}
	}

	// Логируем выигрышный ход
	logStateJSON(playStateBest, "winning_move")

	return playStateBest
}

func getParentState(playState *PlayState) *PlayState {
	return playState.prevState
}
