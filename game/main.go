package game

import (
	"net"
	"strings"
	"bufio"

	log "github.com/inconshreveable/log15"
	"strconv"
	"fmt"
	"os"
)

type Server struct {
	Host	string
	Port	string
}

type GamePlan struct {
	N          int
	Plan       [][]int
	FreeFields map[string][2]int
}

type Game struct {
	PlayersN	int
	GameID 		string
	PlayerName	string
	PlayerID	int
	Plan 		GamePlan
	Conn 		net.Conn
	Reader 		*bufio.Reader
}


func (this *Game) Init(server Server, game_id string, player_name string) {
	this.PlayersN = 2
	this.GameID = game_id
	this.PlayerName = player_name

	var err error
	if this.Conn, err = net.Dial("tcp", server.Host + ":" + server.Port); err != nil {
		panic(err)
	}
	this.Reader = bufio.NewReader(this.Conn)
}

func (this *Game) Send(msg string) {
	log.Debug("Sending some data to server", "data", msg)
	_, err := this.Conn.Write([]byte(msg + "\n"))
	if err != nil {
		panic(err)
	}
}

func (this *Game) Read() string {
	buf, err := this.Reader.ReadBytes('\n')
	if err != nil  {
		if err.Error() != "EOF" {
			panic(err)
		}
	}
	return this.ParseReceived(&buf)
}

func (this *Game) ParseReceived(msg *[]byte) string {
	return strings.Trim(string(*msg), "\r\n ")
}

func (this *GamePlan) Generate(n int) {
	this.N = n
	this.Plan = make([][]int, this.N)
	this.FreeFields = map[string][2]int{}
	for i := range this.Plan {
		this.Plan[i] = make([]int, this.N)
		for j := range this.Plan[i] {
			this.Plan[i][j] = -1
			this.FreeFields[strconv.Itoa(i) + strconv.Itoa(j)] = [2]int{i, j}
		}
	}
}

func (this *GamePlan) Move(player_id, x, y int) {
	delete(this.FreeFields, strconv.Itoa(x) + strconv.Itoa(y))
	this.Plan[y][x] = player_id
}

func (this *GamePlan) RevokeMove(player_id, x, y int) {
	this.FreeFields[strconv.Itoa(x) + strconv.Itoa(y)] = [2]int{x, y}
	this.Plan[y][x] = -1
}

func (this *GamePlan) IsWinning(player_id, last_x, last_y int) (int, bool) {
	var n int = 5
	if this.N < 5 {
		n = this.N
	}
	var count int

	count = 0
	for x := last_x; x >= 0; x-- {
		if this.Plan[last_y][x] == player_id {
			count++
		} else {
			break
		}
	}
	for x := last_x + 1; x < this.N; x++ {
		if this.Plan[last_y][x] == player_id {
			count++
		} else {
			break
		}
	}
	if count >= n {
		return player_id, true
	}

	count = 0
	for y := last_y; y >= 0; y-- {
		if this.Plan[y][last_x] == player_id {
			count++
		} else {
			break
		}
	}
	for y := last_y + 1; y < this.N; y++ {
		if this.Plan[y][last_x] == player_id {
			count++
		} else {
			break
		}
	}
	if count >= n {
		return player_id, true
	}

	count = 0
	for x, y := last_x, last_y; x >= 0 && y >= 0; x, y = x - 1, y - 1 {
		if this.Plan[y][x] == player_id {
			count++
		} else {
			break
		}
	}
	for x, y := last_x + 1, last_y + 1; x < this.N && y < this.N; x, y = x + 1, y + 1 {
		if this.Plan[y][x] == player_id {
			count++
		} else {
			break
		}
	}
	if count >= n {
		return player_id, true
	}

	count = 0
	for x, y := last_x, last_y; x < this.N && y >= 0; x, y = x + 1, y - 1 {
		if this.Plan[y][x] == player_id {
			count++
		} else {
			break
		}
	}
	for x, y := last_x - 1, last_y + 1; x >= 0 && y < this.N; x, y = x - 1, y + 1 {
		if this.Plan[y][x] == player_id {
			count++
		} else {
			break
		}
	}
	if count >= n {
		return player_id, true
	}


	if len(this.FreeFields) == 0 {
		return -1, true
	} else {
		return -2, false
	}
}

func (this *GamePlan) Print() {
	for i := 0; i < this.N; i++ {
		for j := 0; j < this.N; j++ {
			var c string
			if this.Plan[i][j] == -1 {
				c = " "
			} else if this.Plan[i][j] == 0 {
				c = "X"
			} else if this.Plan[i][j] == 1 {
				c = "O"
			}
			fmt.Printf("|%s", c)
		}
		fmt.Println("|")
		os.Stdout.Sync()
	}
}