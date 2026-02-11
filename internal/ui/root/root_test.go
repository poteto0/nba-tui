package root

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"nba-tui/internal/ui/scoreboard"
	"github.com/poteto0/go-nba-sdk/types"
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
	if m.state != scoreboardView {
		t.Errorf("expected initial state scoreboardView, got %v", m.state)
	}

	// Simulate selecting a game
	updatedModel, _ := m.Update(scoreboard.SelectGameMsg{GameId: "123"})
	rootM := updatedModel.(Model)

	if rootM.state != detailView {
		t.Errorf("expected state detailView after SelectGameMsg, got %v", rootM.state)
	}

	if rootM.gameID != "123" {
		t.Errorf("expected gameID 123, got %s", rootM.gameID)
	}
}

func TestRootModel_BackToScoreboard(t *testing.T) {
	client := &mockClient{}
	m := NewModel(client)
	m.state = detailView

	// Pressing 'esc' in detail view should go back to scoreboard
	updatedModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	rootM := updatedModel.(Model)

	if rootM.state != scoreboardView {
		t.Errorf("expected state scoreboardView after Esc, got %v", rootM.state)
	}
}
