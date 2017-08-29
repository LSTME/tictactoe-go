const net = require('net');
const process = require('process');
const readline = require('readline');

class Game {
    constructor(server, port, game_id, player_name) {
        this.ready = false
        this.server = server;
        this.port = port;
        this.game_id = game_id;
        this.player_id = -1;
        this.player_name = player_name;
        this.map = [];

        this.socket = new net.Socket()
        this.socket.connect({
            host: this.server,
            port: this.port
        }, () => {
            console.info('Client connected to service', this.server, this.port)
            this.sign_in()
        })
        this.socket.on('data', (data) => {
            this.process_data.call(this,  data)
        })
    }

    send_command(data) {
        console.log('SENDING', data)
        this.socket.write(data.trim() + "\n")
    }

    get(x, y) {
        return this.map[y][x]
    }

    process_data(data) {
        var commands = data.toString().trim().split(' ')
        console.log('RECEIVED', commands)
        switch (commands[0]) {
            case 'OK':
                commands[2] = parseInt(commands[2], 10);
                this.ready = true;
                this.map = [];
                this.player_id = parseInt(commands[3], 10);
                for(var i = 0; i < commands[2]; i++) {
                    this.map[i] = new Array(commands[2])
                    for(var j = 0; j < commands[2]; j++) {
                        this.map[i][j] = -1;
                    }
                }
                break;
            case 'MOVE':
                commands[1] = parseInt(commands[1], 10);
                commands[2] = parseInt(commands[2], 10);
                if(commands[1] != -1) {
                    this.map[commands[2]][commands[1]] = this.player_id == 0 ? 1 : 0;
                }
                var pos = this.compute_move(commands[1], commands[2]);
                this.move(pos[0], pos[1]);

                this.print();
                break;
            case 'GAMEEND':
                commands[1] = parseInt(commands[1], 10);
                if (commands[1] == -2) {
                    console.info('Opponent disconnected')
                } else if (commands[1] == -1) {
                    console.info('It\'s a draw')
                } else {
                    console.info((commands[1] == this.player_id ? 'You' : 'Opponent') + ' won!');
                }
                break;
            case 'CANNOT':
                this.shit_went_wrong();
                break;
            case 'ERROR':
                this.shit_went_wrong();
                break;
            default:

        }
    }

    compute_move(game, get, x, y) {
        return [x, y]
    }

    print() {
        for (var n of this.map) {
            var row = []
            for(var x of n) {
                row.push(x == -1 ? ' ' : (x == 0 ? 'O' : 'X'));
            }
            console.log('|' + row.join('|') + '|');
        }
    }

    sign_in() {
        this.send_command("HELO " + this.game_id + " " + this.player_name)
    }

    move(x, y) {
        this.map[y][x] = this.player_id
        this.send_command("MOVE " + x + " " + y);
    }

    shit_went_wrong() {
        console.error('Shit went wrong');
        process.exit(-1)
    }
}

exports.Game = Game
