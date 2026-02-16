package root

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/poteto0/go-nba-sdk/types"
	"github.com/stretchr/testify/assert"
	"nba-tui/internal/ui/game_detail"
	"nba-tui/internal/ui/scoreboard"
)

type mockClient struct{}

func (m *mockClient) GetScoreboard() ([]types.Game, error) {
	return []types.Game{{GameId: "123"}}, nil
}
func (m *mockClient) GetBoxScore(gameID string) (types.LiveBoxScoreResponse, error) {
	return types.LiveBoxScoreResponse{Game: types.Game{GameId: gameID}}, nil
}
func (m *mockClient) GetPlayByPlay(gameID string) (types.LivePlayByPlayResponse, error) {
	return types.LivePlayByPlayResponse{}, nil
}

func TestRootModel_Transition(t *testing.T) {
	client := &mockClient{}
	m := NewModel(client)

	// Initially should be scoreboard
	assert.Equal(t, scoreboardView, m.state)

	// Simulate selecting a game
	updatedModel, _ := m.Update(scoreboard.SelectGameMsg{GameId: "123"})
	rootM := updatedModel.(Model)

	assert.Equal(t, detailView, rootM.state)
	assert.Equal(t, "123", rootM.gameID)
}

func TestRootModel_BackToScoreboard(t *testing.T) {

	client := &mockClient{}

	m := NewModel(client)

	m.state = detailView

	// Pressing 'esc' in detail view should go back to scoreboard

	updatedModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyEsc})

	rootM := updatedModel.(Model)

	assert.Equal(t, scoreboardView, rootM.state)

	// Test backspace too

	m.state = detailView

	updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyBackspace})

	rootM = updatedModel.(Model)

	assert.Equal(t, scoreboardView, rootM.state)

}

func TestRootModel_WindowSize(t *testing.T) {

	client := &mockClient{}

	m := NewModel(client)

	updatedModel, _ := m.Update(tea.WindowSizeMsg{Width: 100, Height: 50})

	rootM := updatedModel.(Model)

	assert.Equal(t, 100, rootM.width)

	assert.Equal(t, 50, rootM.height)

}

func TestRootModel_View(t *testing.T) {

	client := &mockClient{}

	m := NewModel(client)

	// Scoreboard view

	m.state = scoreboardView

	view := m.View()

	assert.Contains(t, view, "Loading") // scoreboard initial view with no games

	// Detail view

	m.state = detailView

	m.detailModel = game_detail.New(client, "123")

	view = m.View()

	assert.Contains(t, view, "Loading") // game_detail initial view

}

func TestRootModel_Init(t *testing.T) {

	client := &mockClient{}

	m := NewModel(client)

	cmd := m.Init()

	assert.NotNil(t, cmd)

}

func TestRootModel_UpdateDelegation(t *testing.T) {

	client := &mockClient{}

	m := NewModel(client)

	// Test delegation to scoreboard

	m.state = scoreboardView

	// Send a key msg that scoreboard handles (e.g. 'j' for down)

	updatedModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})

	rootM := updatedModel.(Model)

	assert.Equal(t, 0, rootM.scoreboardModel.Focus) // No games, so focus stays 0

	// Test delegation to detail

	m.state = detailView

	m.detailModel = game_detail.New(client, "123")

	// Send a key msg that detail handles

	updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})

	rootM = updatedModel.(Model)

	assert.NotNil(t, rootM.detailModel)

}
