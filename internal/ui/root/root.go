package root

import (
	tea "github.com/charmbracelet/bubbletea"
	"nba-tui/internal/ui/game_detail"
	"nba-tui/internal/ui/scoreboard"
	"github.com/poteto0/go-nba-sdk/types"
)

type state int

const (
	scoreboardView state = iota
	detailView
)

type Client interface {
	GetScoreboard() ([]types.Game, error)
	GetBoxScore(gameID string) (types.LiveBoxScoreResponse, error)
	GetPlayByPlay(gameID string) (types.LivePlayByPlayResponse, error)
}

type Model struct {
	client          Client
	scoreboardModel scoreboard.Model
	detailModel     game_detail.Model
	state           state
	gameID          string
	width           int
	height          int
}

func NewModel(client Client) Model {
	return Model{
		client:          client,
		scoreboardModel: scoreboard.NewModel(client),
		state:           scoreboardView,
	}
}

func (m Model) Init() tea.Cmd {
	return m.scoreboardModel.Init()
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case scoreboard.SelectGameMsg:
		m.state = detailView
		m.gameID = msg.GameId
		m.detailModel = game_detail.New(m.client, m.gameID)
		// Initialize with current width/height
		dm, _ := m.detailModel.Update(tea.WindowSizeMsg{Width: m.width, Height: m.height})
		m.detailModel = dm.(game_detail.Model)
		return m, m.detailModel.Init()

	case tea.KeyMsg:
		if m.state == detailView && (msg.String() == "esc" || msg.String() == "backspace") {
			m.state = scoreboardView
			return m, nil
		}
	}

	if m.state == scoreboardView {
		var newModel tea.Model
		newModel, cmd = m.scoreboardModel.Update(msg)
		m.scoreboardModel = newModel.(scoreboard.Model)
	} else {
		var newModel tea.Model
		newModel, cmd = m.detailModel.Update(msg)
		m.detailModel = newModel.(game_detail.Model)
	}

	return m, cmd
}

func (m Model) View() string {
	if m.state == scoreboardView {
		return m.scoreboardModel.View()
	}
	return m.detailModel.View()
}
