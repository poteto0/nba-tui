package scoreboard

import (
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
	"github.com/poteto0/go-nba-sdk/types"
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
		if !strings.Contains(view, "Final") {
			t.Errorf("expected view to contain 'Final', got %s", view)
		}
		if !strings.Contains(view, "POR") || !strings.Contains(view, "DEN") {
			t.Errorf("expected view to contain team codes, got %s", view)
		}
		if !strings.Contains(view, "103") || !strings.Contains(view, "102") {
			t.Errorf("expected view to contain scores, got %s", view)
		}
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
		if !strings.Contains(view, "\x1b[1mPOR\x1b[0m") {
			t.Errorf("expected POR to be bolded, got %q", view)
		}
		if !strings.Contains(view, "\x1b[1m103\x1b[0m") {
			t.Errorf("expected 103 to be bolded, got %q", view)
		}
	})

	t.Run("displays help text", func(t *testing.T) {
		m := NewModel(&mockClient{})
		view := m.View()
		if !strings.Contains(view, "move: <hjkli←↓↑→ >, <ctrl+w>: watch (browser), q/esc: quit") {
			t.Errorf("expected view to contain help text, got %q", view)
		}
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
		if !strings.Contains(view, " 9 ") {
			t.Errorf("expected 1-digit score to be centered with spaces, got view: %s", view)
		}
		// 2 digits: " 99" (space, 99)
		if !strings.Contains(view, " 99") {
			t.Errorf("expected 2-digit score to be left-padded with one space, got view: %s", view)
		}
	})

	t.Run("displays last updated time", func(t *testing.T) {
		m := NewModel(&mockClient{})
		// Manually set LastUpdated to a known time
		fixedTime := time.Date(2023, 10, 27, 10, 0, 0, 0, time.Local)
		m.LastUpdated = fixedTime

		view := m.View()
		expectedTimeStr := fixedTime.Format(time.RFC1123)
		if !strings.Contains(view, expectedTimeStr) {
			t.Errorf("expected view to contain last updated time %q, got %q", expectedTimeStr, view)
		}
	})

	t.Run("calculates columns and handles grid navigation", func(t *testing.T) {
		// Create 4 dummy games
		games := make([]types.Game, 4)
		m := NewModel(&mockClient{games: games})
		m.Games = games

		// Simulate window size that fits 2 columns
		// Assuming a box width around 16-20 chars.
		// Let's say 40 width is enough for 2 columns.
		m.Width = 40
		m.calculateColumns() // Helper to trigger logic if extracted

		if m.Columns != 2 {
			t.Errorf("expected 2 columns for width 40, got %d", m.Columns)
		}

		// Initial focus 0 (0,0)
		// Move Right -> 1 (0,1)
		m, _ = updateModel(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("l")})
		if m.Focus != 1 {
			t.Errorf("expected focus 1 after moving right, got %d", m.Focus)
		}

		// Move Down -> 3 (1,1) (1 + 2 columns)
		m, _ = updateModel(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
		if m.Focus != 3 {
			t.Errorf("expected focus 3 after moving down from 1, got %d", m.Focus)
		}

		// Move Up -> 1 (0,1) (3 - 2 columns)
		m, _ = updateModel(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("k")})
		if m.Focus != 1 {
			t.Errorf("expected focus 1 after moving up from 3, got %d", m.Focus)
		}

		// Move Left -> 0
		m, _ = updateModel(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("h")})
		if m.Focus != 0 {
			t.Errorf("expected focus 0 after moving left, got %d", m.Focus)
		}
	})

	t.Run("handles tickMsg", func(t *testing.T) {
		m := NewModel(&mockClient{})
		_, cmd := m.Update(tickMsg{})
		if cmd == nil {
			t.Error("expected command to be returned on tickMsg")
		}
	})
}

func updateModel(m Model, msg tea.Msg) (Model, tea.Cmd) {
	newM, cmd := m.Update(msg)
	return newM.(Model), cmd
}
