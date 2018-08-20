package main

import (
	"os"

	"git.tumeo.eu/lstme/tictactoe-client/game"
	"git.tumeo.eu/lstme/tictactoe-client/mover"
)

func main() {
	hostname := os.Getenv("GAME_HOST")
	if hostname == "" {
		hostname = "localhost"
	}
    port := os.Getenv("GAME_PORT")
	if port == "" {
		port ="32768"
	}
    game_id := os.Getenv("GAME_ID")
	if game_id == "" {
		game_id = "game123"
	}
    player := os.Getenv("GAME_PLAYER")
    if player == "" {
        player = "bob"
    }


	my_game := new(game.Game)
	my_mover := new(mover.Mover)

	my_game.Init(game.Server{
		Host: hostname,
		Port: port,
	}, game_id, player)
	my_mover.Init(my_game)
}
