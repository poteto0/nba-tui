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
<esc>: back, <hjkli←↓↑→ >: move, <enter>: window, <ctrl+w> watch (browser)
view
<alt+d>: default, <alt+l>: line
┌─────────────┐ ┌─────┐
│          Final           │ │gamelog   │
│ POR (103)  |  DEN (102)  │ │4Q(11:43) | Sharpe(POR) Made Jump Sh
└─────────────┘ │...
┌─────────────┐ │
│      Box Scores          │ │
│PORtable view stats       │ │
│...                       │ │
│...                       │ │
└─────────────┘ │
```

#### forcus window key-bind

##### BoxScore

- move is like vim
- switch is switch team showing box

```bash
$ nba-tui
move
<esc>: back, <hjkli←↓↑→ >: move(scroll),
op
<ctrl+s> switch
┌─────────────┐ ┌─────────┐
│          Final           │ │gamelog           │
│ POR (103)  |  DEN (102)  │ │4Q(11:43) | Sharpe(POR) Made Jump Sh
└─────────────┘ │...
┌─────────────┐ │
│      Box Scores          │ │
│PORtable view stats       │ │
│...                       │ │
│...                       │ │
└─────────────┘ │
```

## Development

### Language

- go

### Library

- https://github.com/charmbracelet/bubbletea
- https://github.com/poteto0/go-nba-sdk
