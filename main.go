package main

import (
	"os"

	"git.tumeo.eu/lstme/tictactoe-client/game"
	"git.tumeo.eu/lstme/tictactoe-client/mover"
)

func main() {
	hostname := os.Getenv("GAME_HOST")
	port := os.Getenv("GAME_PORT")
	game_id := os.Getenv("GAME_ID")
	if hostname == "" {
		hostname = "localhost"
	}
	if port == "" {
		port ="32768"
	}
	if game_id == "" {
		game_id = "game123"
	}


	my_game := new(game.Game)
	my_mover := new(mover.Mover)

	my_game.Init(game.Server{
		Host: hostname,
		Port: port,
	}, game_id, "bob")
	my_mover.Init(my_game)
}

