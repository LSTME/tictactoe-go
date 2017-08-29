const { Game } = require('./Game');

class Mover extends Game {
    constructor(server, port, game_id, player_name) {
        super(server, port, game_id, player_name)
    }

    compute_move(opponent_last_x, opponent_last_y) {
        var res;
        while (true) {
            res = [this.get_random_int(0, this.map.length), this.get_random_int(0, this.map.length)];
            if(this.map[res[1]][res[0]] === -1) {
                return res;
            }
        }
        return [opponent_last_x,opponent_last_y]
    }

    get_random_int(min, max) {
        var min = Math.ceil(min);
        var max = Math.floor(max);
        return Math.floor(Math.random() * (max - min)) + min; //The maximum is exclusive and the minimum is inclusive
    }
}

    exports.Mover = Mover
