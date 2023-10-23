package main

import (
	"log"
	"net/http"
	_ "net/http/pprof"
)

func main() {

	// init fields on board that cached and readonly
	initFieldsOnBoard()

	// make handlers for dynamic server content
	//http.HandleFunc("/welcome/", welcomeHandler)
	http.HandleFunc("/", initGameHandler)
	http.HandleFunc("/game/state/", getGameStateHandler)

	// make handler for static content
	fs := http.FileServer(http.Dir("./static"))
	// remember that url for static resources should started from "/", for example: href="/static/stylesheets/checkers.css" or src="/static/javascripts/checkers.js"
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	log.Println("open in browser http://127.0.0.1:8090")
	log.Println("use command to profile: go tool pprof [binary] http://127.0.0.1:8090/debug/pprof/profile")

	err := http.ListenAndServe(":8090", nil)
	if err != nil {
		log.Fatal(err)
	}
}

// func welcomeHandler(w http.ResponseWriter, r *http.Request) {
// 	game := initNewGame()
// 	http.Redirect(w, r, fmt.Sprintf("/game/init/?gameid=%s", game.GameId), http.StatusOK)
// }

// var validPath = regexp.MustCompile("^/(.*)?(.*)$")

// func makeHandler(getfunc func(http.ResponseWriter, *http.Request, string, url.Values)) http.HandlerFunc {
// 	return func(w http.ResponseWriter, r *http.Request) {
// 		m := validPath.FindStringSubmatch(r.URL.Path)
// 		if m == nil {
// 			http.NotFound(w, r)
// 			return
// 		}
// 		url := m[1]
// 		params := r.URL.Query()
// 		getfunc(w, r, url, params)
// 	}
// }
