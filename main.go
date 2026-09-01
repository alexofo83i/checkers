package main

import (
	"log"
	"net/http"
	_ "net/http/pprof"
)

func main() {
	initFieldsOnBoard()

	http.HandleFunc("/", initGameHandler)
	http.HandleFunc("/game/state/", getGameStateHandler)

	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	log.Println("open in browser http://127.0.0.1:8090")
	log.Println("use command to profile: go tool pprof [binary] http://127.0.0.1:8090/debug/pprof/profile")

	err := http.ListenAndServe(":8090", nil)
	if err != nil {
		log.Fatal(err)
	}
}
