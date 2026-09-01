package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/google/uuid"
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
	var gameState Table
	var game *Game

	err := json.NewDecoder(r.Body).Decode(&gameState)
	if err != nil || gameState.Name == "" {
		game = getGameById("nil")
	} else {
		game = getGameById(gameState.Name)
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

func initNewGame() *Game {
	gameId := fmt.Sprint(uuid.New())
	t := *initTable(func(chid string) bool {
		return true
	})
	t.Name = gameId
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
