package main

import "log"

const (
	black byte = 'b'
	white byte = 'w'

	// Базовые веса
	WEIGHT_MATERIAL  = 100 // УВЕЛИЧИВАЕМ - потеря шашки становится критичной
	WEIGHT_POTENTIAL = 5   // Уменьшаем, чтобы не перебивал материал

	// Новые агрессивные веса
	WEIGHT_THREAT        = 15  // Угроза съесть шашку соперника
	WEIGHT_DEFENSE       = 20  // Защита своих шашек (наличие поддержки)
	WEIGHT_ATTACK        = 12  // Атака на шашки соперника
	WEIGHT_VULNERABILITY = -30 // Штраф за уязвимые шашки (которые могут съесть)

	// Позиционные веса
	WEIGHT_CENTER   = 3  // Центр важен, но не критичен
	WEIGHT_MOBILITY = 2  // Мобильность важна для атаки
	WEIGHT_EDGE     = -2 // Штраф за край (там легче съесть)
	WEIGHT_BACK_ROW = 4  // Защита заднего ряда (чтобы не дать пройти в дамки)

	// Специальные веса
	WEIGHT_EXCHANGE = 25 // Бонус за выгодный размен
	WEIGHT_PIN      = 15 // Штраф за "зажатые" шашки (которые не могут ходить)
)

func (playState *PlayState) Cost() int {
	if playState.cost != 0 {
		return playState.cost
	}

	if playState.cntW == 0 && playState.cntB == 0 {
		playState.countPieces()
	}

	// Основные факторы
	materialDiff := playState.calculateMaterialDiff()
	threats := playState.calculateThreats()                 // Угрозы съесть чужие
	vulnerabilities := playState.calculateVulnerabilities() // Уязвимость своих
	defense := playState.calculateDefense()                 // Защищенность своих
	attackPotential := playState.calculateAttackPotential() // Потенциал атаки

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
		// Для белых меняем знаки
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

// calculateMaterialDiff - полная оценка материального преимущества
func (playState *PlayState) calculateMaterialDiff() int {
	// Базовая разница в количестве
	baseDiff := playState.cntW - playState.cntB

	// Позиционная ценность шашек
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

	// Для черных: cntB - cntW + (positionalValueB - positionalValueW)
	// Для белых: cntW - cntB + (positionalValueW - positionalValueB)
	// Используем baseDiff в логировании или для отладки
	_ = baseDiff // можно закомментировать, если не нужна

	return (playState.cntB - playState.cntW) + (positionalValueB - positionalValueW)
}

// getPieceValue - улучшенная версия с использованием color
func (playState *PlayState) getPieceValue(pos string) int {
	ch, exists := playState.c2f[pos]
	if !exists {
		return 0
	}
	color := ch[0]

	value := 1

	// Центральные поля
	if isCenterSquare(pos) {
		value += 2
	}

	// Штраф за край
	if len(pos) > 0 && (pos[0] == 'a' || pos[0] == 'h') {
		value -= 1
	}

	// Бонус за возможность атаки
	if playState.canAttackFrom(pos) {
		value += 2
	}

	// Бонус за близость к превращению в дамку
	row := pos[1]
	if color == white {
		// Белые стремятся к 8-му ряду
		switch row {
		case '7':
			value += 3
		case '6':
			value += 2
		case '5':
			value += 1
		}

		// Защита начальных шашек
		if row == '1' && (pos[0] == 'b' || pos[0] == 'g') {
			value += 1 // важные центральные шашки в начале
		}
	} else { // black
		// Черные стремятся к 1-му ряду
		switch row {
		case '2':
			value += 3
		case '3':
			value += 2
		case '4':
			value += 1
		}

		// Защита черных шашек
		if row == '8' && (pos[0] == 'b' || pos[0] == 'g') {
			value += 1
		}
	}

	return value
}

// Угрозы: сколько шашек соперника можно съесть
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

// Подсчет количества атак с позиции
func (playState *PlayState) countAttacksFrom(pos string) int {
	attacks := 0
	color := playState.c2f[pos][0]
	opponent := white
	if color == white {
		opponent = black
	}

	// Направления для атаки (вперед и назад для простых шашек)
	directions := []struct{ dx, dy int }{
		{-1, 1}, {1, 1}, // вперед
		{-1, -1}, {1, -1}, // назад (для дамок, но пока для всех)
	}

	for _, dir := range directions {
		// Позиция потенциальной жертвы
		victimPos := shiftPos(pos, dir.dx, dir.dy)
		if victimPos == "" {
			continue
		}

		// Позиция для прыжка
		jumpPos := shiftPos(victimPos, dir.dx, dir.dy)
		if jumpPos == "" {
			continue
		}

		// Проверяем, есть ли шашка соперника и свободно ли место за ней
		if victim, exists := playState.c2f[victimPos]; exists && victim[0] == opponent {
			if _, exists := playState.c2f[jumpPos]; !exists {
				attacks++
			}
		}
	}

	return attacks
}

// Уязвимость: сколько способов съесть наши шашки
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

// Подсчет уязвимости конкретной шашки
func (playState *PlayState) countVulnerabilityAt(pos string) int {
	color := playState.c2f[pos][0]
	opponent := white
	if color == white {
		opponent = black
	}

	vulnerabilities := 0

	// Проверяем все возможные атаки со стороны соперника
	directions := []struct{ dx, dy int }{
		{-1, 1}, {1, 1}, // атака спереди
		{-1, -1}, {1, -1}, // атака сзади
	}

	for _, dir := range directions {
		// Откуда может прилететь атака
		attackerPos := shiftPos(pos, -dir.dx, -dir.dy)
		if attackerPos == "" {
			continue
		}

		// Проверяем, есть ли атакующий
		if attacker, exists := playState.c2f[attackerPos]; exists && attacker[0] == opponent {
			// Проверяем, может ли он съесть
			if playState.canAttackFrom(attackerPos) {
				vulnerabilities++
			}
		}
	}

	return vulnerabilities
}

// Защищенность: сколько союзников защищают шашку
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

// Подсчет защитников для позиции
func (playState *PlayState) countDefenders(pos string) int {
	color := playState.c2f[pos][0]
	defenders := 0

	// Проверяем диагонали для поиска защитников
	directions := []struct{ dx, dy int }{
		{-1, -1}, {1, -1}, // сзади-сбоку
		{-1, 1}, {1, 1}, // спереди-сбоку
	}

	for _, dir := range directions {
		defenderPos := shiftPos(pos, -dir.dx, -dir.dy)
		if defenderPos == "" {
			continue
		}

		if defender, exists := playState.c2f[defenderPos]; exists && defender[0] == color {
			// Проверяем, может ли защитник прикрыть (находится на одной диагонали)
			defenders++
		}
	}

	return defenders
}

// Потенциал атаки (шашки, которые могут атаковать в следующем ходу)
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

// Может ли шашка с этой позиции атаковать
func (playState *PlayState) canAttackFrom(pos string) bool {
	return playState.countAttacksFrom(pos) > 0
}

// Штраф за "зажатые" шашки (которые не могут ходить)
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

// Есть ли у шашки хоть какой-то ход
func (playState *PlayState) hasAnyMove(pos string) bool {
	ch, exists := playState.c2f[pos]
	if !exists {
		return false
	}
	color := ch[0]

	// Направления зависят от цвета
	var directions []struct{ dx, dy int }

	if color == white {
		// Белые ходят вверх
		directions = []struct{ dx, dy int }{
			{-1, 1}, // влево-вверх
			{1, 1},  // вправо-вверх
		}
	} else {
		// Черные ходят вниз
		directions = []struct{ dx, dy int }{
			{-1, -1}, // влево-вниз
			{1, -1},  // вправо-вниз
		}
	}

	// Если шашка может атаковать, это уже ход
	if playState.canAttackFrom(pos) {
		return true
	}

	// Проверяем простые ходы
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

// Вспомогательные функции
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

// countPieces - подсчет количества шашек
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

// calculateCenterControl - расчет контроля центра
func (playState *PlayState) calculateCenterControl() int {
	centerW := 0
	centerB := 0

	// Центральные поля
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

// calculateMobility - расчет мобильности
func (playState *PlayState) calculateMobility() int {
	mobilityW := 0
	mobilityB := 0

	// Простой подсчет мобильности без генерации всех ходов
	for pos, ch := range playState.c2f {
		if ch[0] == white {
			mobilityW += playState.countSimpleMoves(pos)
		} else {
			mobilityB += playState.countSimpleMoves(pos)
		}
	}

	return mobilityW - mobilityB
}

// countSimpleMoves - подсчет простых ходов для шашки (без учета взятий)
func (playState *PlayState) countSimpleMoves(pos string) int {
	color := playState.c2f[pos][0]
	moves := 0

	// Направления для простых ходов
	var directions []struct{ dx, dy int }

	if color == white {
		// Белые ходят вверх (увеличение номера ряда)
		directions = []struct{ dx, dy int }{
			{-1, 1}, // влево-вверх
			{1, 1},  // вправо-вверх
		}
	} else {
		// Черные ходят вниз (уменьшение номера ряда)
		directions = []struct{ dx, dy int }{
			{-1, -1}, // влево-вниз
			{1, -1},  // вправо-вниз
		}
	}

	for _, dir := range directions {
		newPos := shiftPos(pos, dir.dx, dir.dy)
		if newPos != "" {
			// Проверяем, свободно ли поле
			if _, exists := playState.c2f[newPos]; !exists {
				moves++
			}
		}
	}

	return moves
}

// calculateEdgePenalty - штраф за нахождение на краю
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

// calculateBackRowBonus - бонус за шашки на последнем ряду
func (playState *PlayState) calculateBackRowBonus() int {
	backRowW := 0 // белые на 1-м ряду
	backRowB := 0 // черные на 8-м ряду

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

// --- 8< -------------------------------------------------------------------------

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
