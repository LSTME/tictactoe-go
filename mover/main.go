package mover

import (
	"strings"
	"strconv"
	"reflect"
	"math"
	"math/rand"

	log "github.com/inconshreveable/log15"

	"git.tumeo.eu/lstme/tictactoe-client/game"
	"encoding/gob"
	"bytes"
	"sync"
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
	var bestScore = math.Inf(-1)
	log.Debug("Calculating master move", "free_fields", this.game.Plan.FreeFields)

	type ScoreMessage struct {
		Score float64
		Move [2]int
	}

	var (
		waiter_workers sync.WaitGroup
		mutex sync.Mutex
		score_channel = make(chan ScoreMessage, 16)
		possibilities = map[string][2]int{}
	)
	for k,v := range this.game.Plan.FreeFields {
		possibilities[k] = v
	}
	for _, pos := range possibilities {
		var buf bytes.Buffer
		var plan game.GamePlan
		enc := gob.NewEncoder(&buf)
		dec := gob.NewDecoder(&buf)
		err := enc.Encode(this.game.Plan)
		err = dec.Decode(&plan)
		if err != nil {
			panic(err)
		}
		waiter_workers.Add(1)

		go func(plan game.GamePlan, pos [2]int) {
			plan.Move(this.game.PlayerID, pos[0], pos[1])
			score := this.minimaxValue(&plan, this.game.PlayersN - 1 - this.game.PlayerID, 0, pos[0], pos[1], math.Inf(-1), math.Inf(+1))
			log.Debug("score for one of the free positions", "pos", pos, "score", score)
			plan.RevokeMove(this.game.PlayerID, pos[0], pos[1])
			score_channel <- ScoreMessage{
				Score: score,
				Move: pos,
			}
			mutex.Lock()
			if score > bestScore {
				bestScore = score
				bestMove = pos
			}
			mutex.Unlock()
			waiter_workers.Done()
		}(plan, pos)
	}
	waiter_workers.Wait()
	close(score_channel)
	log.Debug("After master move", "free_fields", this.game.Plan.FreeFields)
	return bestMove[0], bestMove[1]
}

func (this *Mover) minimaxValue(plan *game.GamePlan, player_on_move, depth, opponent_x, opponent_y int, alpha, beta float64) float64 {
	if winner, won := plan.IsWinning(this.game.PlayersN - 1 - player_on_move, opponent_x, opponent_y); won {
		if winner == this.game.PlayerID {
			return float64(100 - depth)
		} else if winner >= 0 {
			return float64(depth - 100)
		} else {
			return 0
		}
	}

	var stateScore float64
	if player_on_move == this.game.PlayerID {
		stateScore = math.Inf(-1)
	} else {
		stateScore = math.Inf(+1)
	}

	var possibilities = map[string][2]int{}
	for k,v := range plan.FreeFields {
		possibilities[k] = v
	}
	for _, pos := range possibilities {
		plan.Move(player_on_move, pos[0], pos[1])
		score := this.minimaxValue(plan, this.game.PlayersN - 1 - player_on_move, depth + 1, pos[0], pos[1], alpha, beta)
		plan.RevokeMove(player_on_move, pos[0], pos[1])

		if player_on_move == this.game.PlayerID{
			stateScore = math.Max(stateScore, score)
			alpha = math.Max(alpha, stateScore)
		} else {
			stateScore = math.Min(stateScore, score)
			beta = math.Min(beta, stateScore)
		}

		if beta <= alpha {
			break
		}
	}

	return stateScore + 1
}