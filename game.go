package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"sync"
)

type Games struct {
	mx    sync.RWMutex
	games map[string]*Game
}

type Game struct {
	GameId string
	State  Table
}

var gamesStore = new(Games).Init()

func (g *Games) Init() *Games {
	g.games = make(map[string]*Game)
	return g
}

func (g *Games) GetGame(gameId string) (*Game, bool) {
	g.mx.RLock()
	defer g.mx.RUnlock()

	game, ok := g.games[gameId]
	return game, ok
}

func (g *Games) StoreGame(game *Game) {
	g.mx.Lock()
	defer g.mx.Unlock()
	g.games[game.GameId] = game
}

func getGameStateHandler(wr http.ResponseWriter, r *http.Request) {
	// decode input or return error
	var gameState Table
	var game *Game

	// body, _ := ioutil.ReadAll(r.Body)
	// err := json.Unmarshal(body, &gameState)
	err := json.NewDecoder(r.Body).Decode(&gameState)
	if err != nil || gameState.Name == "" {
		game = getGameById("nil")
	} else {
		// just read state
		game = getGameById(gameState.Name)
		// if whodo was not changed then play
		if gameState.Whodo != "" {
			game.State = gameState
			playWithMe(game)
		}
	}
	renderJson(wr, game.State)
}

func getGameById(gameId string) *Game {
	game, exist := gamesStore.GetGame(gameId)
	if !exist {
		game = initNewGame()
	}
	return game
}

func initGameHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		return
	}
	game := getGameById("nil")
	renderTemplate(w, table_html, game.State)
}

// var partState = map[string]struct{}{"b1": {}, "b2": {}, "w9": {}, "w10": {}}

func initNewGame() *Game {
	var t Table

	//jsonDebug := `{"n":"9ac763f4-7bc5-4b62-a6dd-d42bfff6ef48","trs":[{"th":8,"tds":[{"id":"8a","cl":"white","ch":{"id":""}},{"id":"8b","cl":"square black","ch":{"id":"b1"}},{"id":"8c","cl":"white","ch":{"id":""}},{"id":"8d","cl":"square black","ch":{"id":"b2"}},{"id":"8e","cl":"white","ch":{"id":""}},{"id":"8f","cl":"square black","ch":{"id":"b3"}},{"id":"8g","cl":"white","ch":{"id":""}},{"id":"8h","cl":"square black","ch":{"id":"b4"}}]},{"th":7,"tds":[{"id":"7a","cl":"square black","ch":{"id":"b5"}},{"id":"7b","cl":"white","ch":{"id":""}},{"id":"7c","cl":"square black","ch":{"id":""}},{"id":"7d","cl":"white","ch":{"id":""}},{"id":"7e","cl":"square black","ch":{"id":"b7"}},{"id":"7f","cl":"white","ch":{"id":""}},{"id":"7g","cl":"square black","ch":{"id":"b8"}},{"id":"7h","cl":"white","ch":{"id":""}}]},{"th":6,"tds":[{"id":"6a","cl":"white","ch":{"id":""}},{"id":"6b","cl":"square black","ch":{"id":"b9"}},{"id":"6c","cl":"white","ch":{"id":""}},{"id":"6d","cl":"square black","ch":{"id":""}},{"id":"6e","cl":"white","ch":{"id":""}},{"id":"6f","cl":"square black","ch":{"id":""}},{"id":"6g","cl":"white","ch":{"id":""}},{"id":"6h","cl":"square black","ch":{"id":"b12"}}]},{"th":5,"tds":[{"id":"5a","cl":"square black","ch":{"id":""}},{"id":"5b","cl":"white","ch":{"id":""}},{"id":"5c","cl":"square black","ch":{"id":""}},{"id":"5d","cl":"white","ch":{"id":""}},{"id":"5e","cl":"square black","ch":{"id":"b6"}},{"id":"5f","cl":"white","ch":{"id":""}},{"id":"5g","cl":"square black","ch":{"id":"b11"}},{"id":"5h","cl":"white","ch":{"id":""}}]},{"th":4,"tds":[{"id":"4a","cl":"white","ch":{"id":""}},{"id":"4b","cl":"square black","ch":{"id":""}},{"id":"4c","cl":"white","ch":{"id":""}},{"id":"4d","cl":"square black","ch":{"id":""}},{"id":"4e","cl":"white","ch":{"id":""}},{"id":"4f","cl":"square black","ch":{"id":""}},{"id":"4g","cl":"white","ch":{"id":""}},{"id":"4h","cl":"square black","ch":{"id":"w4"}}]},{"th":3,"tds":[{"id":"3a","cl":"square black","ch":{"id":"w1"}},{"id":"3b","cl":"white","ch":{"id":""}},{"id":"3c","cl":"square black","ch":{"id":""}},{"id":"3d","cl":"white","ch":{"id":""}},{"id":"3e","cl":"square black","ch":{"id":"w3"}},{"id":"3f","cl":"white","ch":{"id":""}},{"id":"3g","cl":"square black","ch":{"id":""}},{"id":"3h","cl":"white","ch":{"id":""}}]},{"th":2,"tds":[{"id":"2a","cl":"white","ch":{"id":""}},{"id":"2b","cl":"square black","ch":{"id":"w5"}},{"id":"2c","cl":"white","ch":{"id":""}},{"id":"2d","cl":"square black","ch":{"id":"w6"}},{"id":"2e","cl":"white","ch":{"id":""}},{"id":"2f","cl":"square black","ch":{"id":"w7"}},{"id":"2g","cl":"white","ch":{"id":""}},{"id":"2h","cl":"square black","ch":{"id":"w8"}}]},{"th":1,"tds":[{"id":"1a","cl":"square black","ch":{"id":"w9"}},{"id":"1b","cl":"white","ch":{"id":""}},{"id":"1c","cl":"square black","ch":{"id":"w10"}},{"id":"1d","cl":"white","ch":{"id":""}},{"id":"1e","cl":"square black","ch":{"id":"w11"}},{"id":"1f","cl":"white","ch":{"id":""}},{"id":"1g","cl":"square black","ch":{"id":"w12"}},{"id":"1h","cl":"white","ch":{"id":""}}]}],"whodo":"checkers-black","next":[{"ch":"w6","steps":["3c"],"kicks":[]},{"ch":"w8","steps":["3g"],"kicks":[]},{"ch":"w5","steps":["3c"],"kicks":[]},{"ch":"w1","steps":["4b"],"kicks":[]},{"ch":"w4","steps":[],"kicks":["b11-6f,b6-4d"]},{"ch":"w7","steps":["3g"],"kicks":[]},{"ch":"w3","steps":["4f","4d"],"kicks":[]}]}`
	jsonDebug := `{"n":"9ac763f4-7bc5-4b62-a6dd-d42bfff6ef48","trs":[{"th":8,"tds":[{"id":"8a","cl":"white","ch":{"id":""}},{"id":"8b","cl":"square black","ch":{"id":"b1"}},{"id":"8c","cl":"white","ch":{"id":""}},{"id":"8d","cl":"square black","ch":{"id":"b2"}},{"id":"8e","cl":"white","ch":{"id":""}},{"id":"8f","cl":"square black","ch":{"id":"b3"}},{"id":"8g","cl":"white","ch":{"id":""}},{"id":"8h","cl":"square black","ch":{"id":"b4"}}]},{"th":7,"tds":[{"id":"7a","cl":"square black","ch":{"id":""}},{"id":"7b","cl":"white","ch":{"id":""}},{"id":"7c","cl":"square black","ch":{"id":""}},{"id":"7d","cl":"white","ch":{"id":""}},{"id":"7e","cl":"square black","ch":{"id":"b7"}},{"id":"7f","cl":"white","ch":{"id":""}},{"id":"7g","cl":"square black","ch":{"id":"b8"}},{"id":"7h","cl":"white","ch":{"id":""}}]},{"th":6,"tds":[{"id":"6a","cl":"white","ch":{"id":""}},{"id":"6b","cl":"square black","ch":{"id":"b5"}},{"id":"6c","cl":"white","ch":{"id":""}},{"id":"6d","cl":"square black","ch":{"id":""}},{"id":"6e","cl":"white","ch":{"id":""}},{"id":"6f","cl":"square black","ch":{"id":""}},{"id":"6g","cl":"white","ch":{"id":""}},{"id":"6h","cl":"square black","ch":{"id":"b12"}}]},{"th":5,"tds":[{"id":"5a","cl":"square black","ch":{"id":"b9"}},{"id":"5b","cl":"white","ch":{"id":""}},{"id":"5c","cl":"square black","ch":{"id":""}},{"id":"5d","cl":"white","ch":{"id":""}},{"id":"5e","cl":"square black","ch":{"id":""}},{"id":"5f","cl":"white","ch":{"id":""}},{"id":"5g","cl":"square black","ch":{"id":""}},{"id":"5h","cl":"white","ch":{"id":""}}]},{"th":4,"tds":[{"id":"4a","cl":"white","ch":{"id":""}},{"id":"4b","cl":"square black","ch":{"id":"w1"}},{"id":"4c","cl":"white","ch":{"id":""}},{"id":"4d","cl":"square black","ch":{"id":"w4"}},{"id":"4e","cl":"white","ch":{"id":""}},{"id":"4f","cl":"square black","ch":{"id":""}},{"id":"4g","cl":"white","ch":{"id":""}},{"id":"4h","cl":"square black","ch":{"id":""}}]},{"th":3,"tds":[{"id":"3a","cl":"square black","ch":{"id":""}},{"id":"3b","cl":"white","ch":{"id":""}},{"id":"3c","cl":"square black","ch":{"id":""}},{"id":"3d","cl":"white","ch":{"id":""}},{"id":"3e","cl":"square black","ch":{"id":"w3"}},{"id":"3f","cl":"white","ch":{"id":""}},{"id":"3g","cl":"square black","ch":{"id":""}},{"id":"3h","cl":"white","ch":{"id":""}}]},{"th":2,"tds":[{"id":"2a","cl":"white","ch":{"id":""}},{"id":"2b","cl":"square black","ch":{"id":"w5"}},{"id":"2c","cl":"white","ch":{"id":""}},{"id":"2d","cl":"square black","ch":{"id":"w6"}},{"id":"2e","cl":"white","ch":{"id":""}},{"id":"2f","cl":"square black","ch":{"id":"w7"}},{"id":"2g","cl":"white","ch":{"id":""}},{"id":"2h","cl":"square black","ch":{"id":"w8"}}]},{"th":1,"tds":[{"id":"1a","cl":"square black","ch":{"id":"w9"}},{"id":"1b","cl":"white","ch":{"id":""}},{"id":"1c","cl":"square black","ch":{"id":"w10"}},{"id":"1d","cl":"white","ch":{"id":""}},{"id":"1e","cl":"square black","ch":{"id":"w11"}},{"id":"1f","cl":"white","ch":{"id":""}},{"id":"1g","cl":"square black","ch":{"id":"w12"}},{"id":"1h","cl":"white","ch":{"id":""}}]}],"whodo":"checkers-black","next":[{"ch":"","steps":null,"kicks":null},{"ch":"","steps":null,"kicks":null},{"ch":"","steps":null,"kicks":null},{"ch":"","steps":null,"kicks":null},{"ch":"","steps":null,"kicks":null},{"ch":"","steps":null,"kicks":null},{"ch":"","steps":null,"kicks":null},{"ch":"w5","steps":["3a","3c"],"kicks":[]},{"ch":"w4","steps":["5c","5e"],"kicks":[]},{"ch":"w3","steps":["4f"],"kicks":[]},{"ch":"w7","steps":["3g"],"kicks":[]},{"ch":"w1","steps":["5c"],"kicks":[]},{"ch":"w6","steps":["3c"],"kicks":[]},{"ch":"w8","steps":["3g"],"kicks":[]}]}`
	err := json.NewDecoder(strings.NewReader(jsonDebug)).Decode(&t)
	if err != nil {
		log.Fatal(err)
	}
	gameId := t.Name
	// gameId := fmt.Sprint(uuid.New())
	// t := initTable(func(chid string) bool {
	// 	// _, exist := partState[chid]
	// 	// return exist
	// 	return true
	// },
	// )
	// t.Name = gameId
	game := Game{
		GameId: gameId,
		State:  t,
	}

	playState := game.State.convertGame2PlayState()
	log.Default().Println(playState.ToString())

	fillPlayStateByNextSteps(playState, &game.State, playState.whodo)

	gamesStore.StoreGame(&game)
	return &game
}
