# nba-tui

## NBA Client for Programmer

- 💄 Terminal UI for NBA Score
- 👾 90% VibeCoding
- ✨ Vim-Like KeyBinding
- 🤩 Awesome Example of [go-nba-sdk](https://github.com/poteto0/go-nba-sdk)

## Quick Start

```bash
$ just build
$ ./nba-tui
```

## Options

| Option      | Description                                       | Default | Minimum |
|-------------|---------------------------------------------------|---------|---------|
| `--reload`  | Auto-refresh interval for game data in seconds.   | 30      | 10      |
| `--kawaii`  | Enable kawaii mode with special decorations (on/off). | on    | -       |

## Kawaii Mode

When enabled, special achievements are highlighted with icons:

- `👑`: Triple Double
- `💯`: 5x5 (At least 5 in PTS, REB, AST, STL, BLK)
- `🎯`: Sniper (3PM >= 8 & 3P% >= 50%)
- `👽`: PTS >= 50
- `💪`: REB >= 20
- `🤝`: AST >= 20
- `🛡️`: BLK >= 7
- `🥷🏻`: STL >= 5

Also underlines double-digit stats (PTS/REB/AST) and high defensive stats (STL/BLK > 3).

![score-board](./docs/img/score_board.png)
![game-detail](./docs/img/game_detail.png)
