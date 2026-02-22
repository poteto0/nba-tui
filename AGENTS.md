# nba-tui

NBAのスコアボードを見るためのTUIクライアント

`vim`に近いキーバインドでNBAのスコアボードを見る

## Data

### Design

- focused frame is grean
- Win team's score & name is Bolded

### Value

- home team is left

## View

### score board

top page is score board.
default 30s reloaded.

```bash
$ nba-tui
move
<hjkli←↓↑→ >: move, <enter>: detail, <ctrl+w> watch (browser)
view
<alt+d>: default, <alt+l>: line, <ctrl+s> show key player, <esc>: back(show key)
┌───────────┐ ┌───────────┐
│   Final   │ │ 3Q (8:15) │
│ POR | DEN │ │ LAL | LAC │
│ --------- │ │ --------- │
│ 103 | 102 │ │  91 | 72  │
└───────────┘ └───────────┘
```

#### Show Key Player

```bash
... not decided
```

### Game Detail View

focus & push <enter> to detail view

```bash
$ nba-tui
moves
<esc>: back, <hjkli←↓↑→ >: move, <ctrl+b>: focus box score, <ctrl+l>: focus game log, <ctrl+w> watch (browser)
view
Selected Team: POR
┌─────────────┐ ┌─────────────────────────┐
│          Final           │ │         gamelog         │
│ POR (103)  |  DEN (102)  │ │    1Q | 2Q | 3Q | 4Q     │
└─────────────┘ │4Q(11:43) | Sharpe(POR) Made Jump Sh
┌─────────────┐ │...
│      Box Scores          │ │
│PORtable view stats       │ │
│...                       │ │
│...                       │ │
└─────────────┘ │
```

#### forcus window key-bind

##### BoxScore & GameLog

- focus switch: <ctrl+b> for BoxScore, <ctrl+l> for GameLog
- scroll: <hjkli←↓↑→ > (vim-like)
- horizontal scroll (Box Score): <h/l> or <left/right> arrows
- team switch: <ctrl+s> switch team for both BoxScore and GameLog
- period switch: <ctrl+q> switch period for GameLog (1Q -> 2Q -> ... -> Final -> 1Q)

##### Box Score Columns

- Format: `MIN FGM FGA FG% 3PM 3PA 3P% FTM FTA FT% OREB DREB REB AST STL BLK TO PF PTS +/-`
- `MIN`: First 5 characters (e.g., `36:10`). Shows `-` if length <= 5.
- Alignment: `PLAYER` and `MIN` are left-aligned. All other stats are right-aligned.

##### Game Status Display

- "Not Started": When the game hasn't started yet.
- Overtime: Displayed as `1OT`, `2OT`, etc. (instead of `5Q`, `6Q`).

##### UI Style

- Selected Period in GameLog: Green color + Underline
- Unselected Period in GameLog: Faint color
- Selected Pane: Green border
- Selected Team Display: "Selected Team: <TeamTriCode>" above panes

## Development

### Language

- go

### Library

- https://github.com/charmbracelet/bubbletea
- https://github.com/poteto0/go-nba-sdk

### Mock Mode

Start with mock data for development:

```bash
go run ./cmd/nba-tui/main.go --mock
```
