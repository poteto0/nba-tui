package root

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/poteto0/go-nba-sdk/types"
	"nba-tui/internal/ui/game_detail"
	"nba-tui/internal/ui/scoreboard"
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

// TickMsg is a message that indicates a time-based event, typically used for periodic updates.
type TickMsg time.Time

// tickCmd returns a tea.Cmd that sends a TickMsg after the specified duration.
func tickCmd(interval time.Duration) tea.Cmd {
	return tea.Tick(interval, func(t time.Time) tea.Msg {
		return TickMsg(t)
	})
}

type Model struct {
	client          Client
	scoreboardModel scoreboard.Model
	detailModel     game_detail.Model
	state           state
	gameID          string
	width           int
	height          int
	config          game_detail.Config
	reloadInterval  time.Duration // New field for reload interval
}

func NewModel(client Client, config game_detail.Config, reload int) Model {
	return Model{
		client:          client,
		scoreboardModel: scoreboard.NewModel(client),
		state:           scoreboardView,
		config:          config,
		reloadInterval:  time.Duration(reload) * time.Second,
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(m.scoreboardModel.Init(), tickCmd(m.reloadInterval))
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case scoreboard.SelectGameMsg:
		m.state = detailView
		m.gameID = msg.GameId
		m.detailModel = game_detail.New(m.client, m.gameID, m.config)
		// Initialize with current width/height
		dm, _ := m.detailModel.Update(tea.WindowSizeMsg{Width: m.width, Height: m.height})
		m.detailModel = dm.(game_detail.Model)
		return m, tea.Batch(m.detailModel.Init(), tickCmd(m.reloadInterval))

	case TickMsg:
		if m.state == scoreboardView {
			cmds = append(cmds, m.scoreboardModel.FetchScoreboard())
		} else if m.state == detailView {
			// Ensure detailModel has the latest width/height before refreshing
			dm, _ := m.detailModel.Update(tea.WindowSizeMsg{Width: m.width, Height: m.height})
			m.detailModel = dm.(game_detail.Model)
			cmds = append(cmds, m.detailModel.Init()) // Re-initialize to fetch new data
		}
		cmds = append(cmds, tickCmd(m.reloadInterval)) // Restart the timer
		return m, tea.Batch(cmds...)

	case tea.KeyMsg:
		if m.state == detailView && (msg.String() == "esc" || msg.String() == "backspace") {
			m.state = scoreboardView
			return m, tickCmd(m.reloadInterval)
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
