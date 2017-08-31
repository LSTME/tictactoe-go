package mover

import (
	"strings"
	"strconv"
	"reflect"
	"math/rand"

	log "github.com/inconshreveable/log15"

	"git.tumeo.eu/lstme/tictactoe-client/game"
)

type Mover struct {
	game *game.Game
}

func (this *Mover) Init(my_game *game.Game) {
	this.game = my_game
	this.signIn()
	for {
		data := strings.Split(this.game.Read(), " ")
		log.Debug("Received some data from server", "data", data)
		switch data[0] {
		case "OK":
			if data[1] != this.game.GameID {
				panic("shit went wrong")
			}
			n1, err := strconv.Atoi(data[2])
			n2, err := strconv.Atoi(data[3])
			if err != nil {
				panic("shit went wrong" + err.Error())
			}
			this.game.Plan.Generate(n1)
			this.game.PlayerID = n2

		case "GAMEEND":
			var msg string
			winner_id, _ := strconv.Atoi(data[1])
			switch winner_id {
			case -2:
				msg = "Opponent disconnected"
			case -1:
				msg = "Draw"
			default:
				if winner_id == this.game.PlayerID {
					msg = "You won!"
				} else {
					msg = "Opponent won"
				}
			}

			log.Info(msg)
			return

		case "MOVE":
			x, err := strconv.Atoi(data[1])
			y, err := strconv.Atoi(data[2])
			if err != nil {
				panic(err)
			}

			if x != -1 && y != -1 {
				this.game.Plan.Move(this.game.PlayersN - 1 - this.game.PlayerID, x, y)
			}
			x, y = this.calculateMove(x, y)
			this.game.Plan.Move(this.game.PlayerID, x, y)
			this.game.Send("MOVE " + strconv.Itoa(x) + " " + strconv.Itoa(y))

		default:
			panic("shit went wrong")
		}
	}
}


func (this *Mover) signIn() {
	log.Debug("Signing in")
	this.game.Send("HELO " + this.game.GameID + " " + this.game.PlayerName)
}


func (this *Mover) calculateMove(opponent_x, opponent_y int) (x, y int){
	return this.masterMove(opponent_x, opponent_y)
}

func (this *Mover) blindMove() (x, y int ){
	r := rand.New(rand.NewSource(99))
	keys := reflect.ValueOf(this.game.Plan.FreeFields).MapKeys()
	key := keys[r.Intn(len(keys))]
	xy := this.game.Plan.FreeFields[key.String()]
	return xy[0], xy[1]
}

func (this *Mover) masterMove(opponent_x, opponent_y int) (x, y int) {
	var bestMove [2]int
	var bestScore = -100*this.game.Plan.N*this.game.Plan.N
	log.Debug("Calculating master move", "free_fields", this.game.Plan.FreeFields)
	var possibilities = map[string][2]int{}
	for k,v := range this.game.Plan.FreeFields {
		possibilities[k] = v
	}
	for _, pos := range possibilities {
		this.game.Plan.Move(this.game.PlayerID, pos[0], pos[1])
		score := this.minimaxValue(&this.game.Plan, this.game.PlayersN - 1 - this.game.PlayerID, 0, pos[0], pos[1])
		log.Debug("score for one of the free positions", "pos", pos, "score", score)
		this.game.Plan.RevokeMove(this.game.PlayerID, pos[0], pos[1])
		if score > bestScore {
			bestScore = score
			bestMove = pos
		}
	}
	log.Debug("After master move", "free_fields", this.game.Plan.FreeFields)
	return bestMove[0], bestMove[1]
}

func (this *Mover) minimaxValue(plan *game.GamePlan, player_on_move, depth, opponent_x, opponent_y int) int {
	if winner, won := plan.IsWinning(this.game.PlayersN - 1 - player_on_move, opponent_x, opponent_y); won {
		if winner == this.game.PlayerID {
			return 10 - depth
		} else if winner >= 0 {
			return depth - 10
		} else {
			return 0
		}
	}

	var stateScore int
	if player_on_move == this.game.PlayerID {
		stateScore = -1000000
	} else {
		stateScore = 1000000
	}

	var possibilities = map[string][2]int{}
	for k,v := range this.game.Plan.FreeFields {
		possibilities[k] = v
	}
	for _, pos := range possibilities {
		plan.Move(player_on_move, pos[0], pos[1])
		score := this.minimaxValue(plan, this.game.PlayersN - 1 - player_on_move, depth + 1, pos[0], pos[1])
		plan.RevokeMove(player_on_move, pos[0], pos[1])

		if player_on_move == this.game.PlayerID && score > stateScore {
			stateScore = score
		} else if player_on_move != this.game.PlayerID && score < stateScore {
			stateScore = score
		}
	}

	return stateScore
}