package scoreboard

import (
	"fmt"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
	"github.com/poteto0/go-nba-sdk/types"
	"github.com/stretchr/testify/assert"
)

type mockClient struct {
	games []types.Game
	err   error
}

func (m *mockClient) GetScoreboard() ([]types.Game, error) {
	return m.games, m.err
}

func TestScoreboardView(t *testing.T) {
	lipgloss.SetColorProfile(termenv.ANSI256)

	t.Run("renders a not started game", func(t *testing.T) {
		games := []types.Game{
			{
				HomeTeam:   types.Team{TeamTricode: "POR"},
				AwayTeam:   types.Team{TeamTricode: "DEN"},
				GameStatus: 1,
			},
		}

		m := NewModel(&mockClient{games: games})
		m.Games = games

		view := m.View()
		assert.Contains(t, view, "Not Started")
	})

	t.Run("renders a finished game", func(t *testing.T) {
		games := []types.Game{
			{
				HomeTeam:       types.Team{TeamTricode: "POR", Score: 103},
				AwayTeam:       types.Team{TeamTricode: "DEN", Score: 102},
				GameStatus:     3,
				GameStatusText: "Final",
			},
		}

		m := NewModel(&mockClient{games: games})
		m.Games = games

		view := m.View()
		assert.Contains(t, view, "Final")
		assert.Contains(t, view, "POR")
		assert.Contains(t, view, "DEN")
		assert.Contains(t, view, "103")
		assert.Contains(t, view, "102")
	})

	t.Run("bolds winning team", func(t *testing.T) {
		games := []types.Game{
			{
				HomeTeam: types.Team{TeamTricode: "POR", Score: 103},
				AwayTeam: types.Team{TeamTricode: "DEN", Score: 100},
			},
		}
		m := NewModel(&mockClient{games: games})
		m.Games = games

		view := m.View()
		// Bold style escape code is \x1b[1m
		assert.Contains(t, view, "\x1b[1mPOR\x1b[0m")
		assert.Contains(t, view, "\x1b[1m103\x1b[0m")
	})

	t.Run("displays help text", func(t *testing.T) {
		m := NewModel(&mockClient{})
		view := m.View()
		assert.Contains(t, view, "<hjkli←↓↑→ >: move, <enter>: detail, <ctrl+w>: watch (browser), <q/esc>: quit")
	})

	t.Run("formats scores correctly", func(t *testing.T) {
		games := []types.Game{
			{
				HomeTeam: types.Team{TeamTricode: "A", Score: 9},
				AwayTeam: types.Team{TeamTricode: "B", Score: 99},
			},
			{
				HomeTeam: types.Team{TeamTricode: "C", Score: 100},
				AwayTeam: types.Team{TeamTricode: "D", Score: 5},
			},
		}
		m := NewModel(&mockClient{games: games})
		m.Games = games
		view := m.View()

		// 1 digit: " 9 " (space, 9, space)
		assert.Contains(t, view, " 9 ")
		// 2 digits: " 99" (space, 99)
		assert.Contains(t, view, " 99")
	})

	t.Run("displays last updated time", func(t *testing.T) {
		m := NewModel(&mockClient{})
		// Manually set LastUpdated to a known time
		fixedTime := time.Date(2023, 10, 27, 10, 0, 0, 0, time.Local)
		m.LastUpdated = fixedTime

		view := m.View()
		expectedTimeStr := fixedTime.Format(time.RFC1123)
		assert.Contains(t, view, expectedTimeStr)
	})

	t.Run("calculates columns and handles grid navigation", func(t *testing.T) {
		// Create 4 dummy games
		games := make([]types.Game, 4)
		m := NewModel(&mockClient{games: games})
		m.Games = games

		// Simulate window size that fits 2 columns
		m.Width = 40
		m.calculateColumns()

		assert.Equal(t, 2, m.Columns)

		// Initial focus 0 (0,0)
		// Move Right -> 1 (0,1)
		m, _ = updateModel(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("l")})
		assert.Equal(t, 1, m.Focus)

		// Move Down -> 3 (1,1) (1 + 2 columns)
		m, _ = updateModel(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
		assert.Equal(t, 3, m.Focus)

		// Move Up -> 1 (0,1) (3 - 2 columns)
		m, _ = updateModel(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("k")})
		assert.Equal(t, 1, m.Focus)

		// Move Left -> 0
		m, _ = updateModel(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("h")})
		assert.Equal(t, 0, m.Focus)
	})

	t.Run("selects a game on enter", func(t *testing.T) {
		games := []types.Game{
			{GameId: "123", HomeTeam: types.Team{TeamTricode: "LAL"}},
			{GameId: "456", HomeTeam: types.Team{TeamTricode: "GSW"}},
		}
		m := NewModel(&mockClient{games: games})
		m.Games = games
		m.Focus = 1 // Focus on second game

		// Simulate Enter key
		_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
		assert.NotNil(t, cmd)

		msg := cmd()
		selectMsg, ok := msg.(SelectGameMsg)
		assert.True(t, ok)
		assert.Equal(t, "456", selectMsg.GameId)
	})

	t.Run("handles error message", func(t *testing.T) {
		m := NewModel(&mockClient{})
		err := fmt.Errorf("api error")
		updatedModel, _ := m.Update(err)
		assert.Equal(t, err, updatedModel.(Model).Err)
		assert.Contains(t, updatedModel.View(), "Error: api error")
	})

	t.Run("handles window size message", func(t *testing.T) {
		m := NewModel(&mockClient{})
		updatedModel, _ := m.Update(tea.WindowSizeMsg{Width: 100, Height: 50})
		assert.Equal(t, 100, updatedModel.(Model).Width)
		assert.Equal(t, 50, updatedModel.(Model).Height)
	})

	t.Run("handles quit keys", func(t *testing.T) {
		m := NewModel(&mockClient{})
		keys := []string{"q", "esc", "ctrl+c"}
		for _, k := range keys {
			_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(k)})
			if k == "ctrl+c" {
				_, cmd = m.Update(tea.KeyMsg{Type: tea.KeyCtrlC})
			} else if k == "esc" {
				_, cmd = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
			}
			assert.NotNil(t, cmd)
			// tea.Quit command is not easily inspectable but we expect it.
		}
	})

	t.Run("handles ctrl+w watch key", func(t *testing.T) {
		games := []types.Game{{GameId: "123"}}
		m := NewModel(&mockClient{games: games})
		m.Games = games

		var openedURL string
		m.OpenBrowser = func(url string) error {
			openedURL = url
			return nil
		}

		_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyCtrlW})
		assert.Nil(t, cmd)
		assert.Equal(t, "https://www.nba.com/game/123", openedURL)
	})

	t.Run("renders away team winner", func(t *testing.T) {
		games := []types.Game{
			{
				HomeTeam: types.Team{TeamTricode: "POR", Score: 100},
				AwayTeam: types.Team{TeamTricode: "DEN", Score: 103},
			},
		}
		m := NewModel(&mockClient{games: games})
		m.Games = games
		view := m.View()
		assert.Contains(t, view, "\x1b[1mDEN\x1b[0m")
		assert.Contains(t, view, "\x1b[1m103\x1b[0m")
	})

	t.Run("renders live game status", func(t *testing.T) {
		games := []types.Game{
			{
				HomeTeam:   types.Team{TeamTricode: "POR"},
				AwayTeam:   types.Team{TeamTricode: "DEN"},
				Period:     2,
				GameClock:  "5:00",
				GameStatus: 2,
			},
		}
		m := NewModel(&mockClient{games: games})
		m.Games = games
		view := m.View()
		assert.Contains(t, view, "2Q (5:00)")
	})

	t.Run("navigation boundaries", func(t *testing.T) {
		games := make([]types.Game, 2)
		m := NewModel(&mockClient{games: games})
		m.Games = games
		m.Columns = 1

		// Left boundary
		m.Focus = 0
		m, _ = updateModel(m, tea.KeyMsg{Type: tea.KeyLeft})
		assert.Equal(t, 0, m.Focus)

		// Right boundary
		m.Focus = 1
		m, _ = updateModel(m, tea.KeyMsg{Type: tea.KeyRight})
		assert.Equal(t, 1, m.Focus)

		// Up boundary
		m.Focus = 0
		m, _ = updateModel(m, tea.KeyMsg{Type: tea.KeyUp})
		assert.Equal(t, 0, m.Focus)

		// Down boundary
		m.Focus = 1
		m, _ = updateModel(m, tea.KeyMsg{Type: tea.KeyDown})
		assert.Equal(t, 1, m.Focus)
	})

	t.Run("fetch scoreboard success", func(t *testing.T) {
		games := []types.Game{{GameId: "123"}}
		client := &mockClient{games: games}
		m := NewModel(client)
		cmd := m.FetchScoreboard()
		msg := cmd()
		assert.Equal(t, GotScoreboardMsg{Games: games}, msg)

		updatedModel, _ := m.Update(msg)
		assert.Equal(t, games, updatedModel.(Model).Games)
	})

	t.Run("fetch scoreboard failure", func(t *testing.T) {
		err := fmt.Errorf("network error")
		client := &mockClient{err: err}
		m := NewModel(client)
		cmd := m.FetchScoreboard()
		msg := cmd()
		assert.Equal(t, err, msg)
	})

	t.Run("calculateColumns with zero width", func(t *testing.T) {
		m := NewModel(&mockClient{})
		m.Width = 0
		m.calculateColumns()
		assert.Equal(t, 1, m.Columns)
	})
}

func updateModel(m Model, msg tea.Msg) (Model, tea.Cmd) {
	newM, cmd := m.Update(msg)
	return newM.(Model), cmd
}
