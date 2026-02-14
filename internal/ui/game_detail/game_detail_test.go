package game_detail

import (
	"fmt"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
	"github.com/poteto0/go-nba-sdk/types"
	"github.com/stretchr/testify/assert"
)

func ptr[T any](v T) *T {
	return &v
}

// Mock Client
type mockNbaClient struct {
	boxScore types.LiveBoxScoreResponse
	pbp      types.LivePlayByPlayResponse
	err      error
}

func (m *mockNbaClient) GetBoxScore(gameID string) (types.LiveBoxScoreResponse, error) {
	return m.boxScore, m.err
}

func (m *mockNbaClient) GetPlayByPlay(gameID string) (types.LivePlayByPlayResponse, error) {
	return m.pbp, m.err
}

func TestNew(t *testing.T) {
	client := &mockNbaClient{}
	gameID := "12345"
	m := New(client, gameID)

	// Check if returns a valid Model (tea.Model)
	assert.NotNil(t, m)
}

func TestInit(t *testing.T) {
	client := &mockNbaClient{}
	gameID := "12345"
	m := New(client, gameID)

	cmd := m.Init()
	if cmd == nil {
		t.Error("Init should return a command to fetch data")
	}
}

func TestUpdate_View(t *testing.T) {
	// Arrange
	client := &mockNbaClient{}
	m := New(client, "123")
	m.width = 100
	m.height = 40

	t.Run("update window size", func(t *testing.T) {
		// Arrange
		msg := tea.WindowSizeMsg{
			Width:  200,
			Height: 80,
		}

		// Act
		model, _ := m.Update(msg)

		// Assert
		assert.Equal(t, 200, model.(Model).width)
		assert.Equal(t, 80, model.(Model).height)

		// cleanup
		m.width = 100
		m.height = 40
	})

	t.Run("update box score view", func(t *testing.T) {
		// Arrange
		boxScore := types.LiveBoxScoreResponse{
			Game: types.Game{
				GameId:   "123",
				HomeTeam: types.Team{TeamName: "Lakers", TeamTricode: "LAL"},
				AwayTeam: types.Team{TeamName: "Warriors", TeamTricode: "GSW"},
			},
		}

		// Act
		updatedModel, _ := m.Update(BoxScoreMsg(boxScore))

		view := updatedModel.View()
		assert.Contains(t, view, "LAL")
		assert.Contains(t, view, "GSW")
	})

	t.Run("update playbyplay view", func(t *testing.T) {
		// Arrange
		// Pre-load BoxScore to bypass Loading screen and set HomeTeam ID
		boxScore := types.LiveBoxScoreResponse{
			Game: types.Game{
				GameId:   "123",
				HomeTeam: types.Team{TeamId: 10, TeamTricode: "LAL"},
			},
		}
		modelWithBoxScore, _ := m.Update(BoxScoreMsg(boxScore))

		pbp := types.LivePlayByPlayResponse{
			Game: types.PlayByPlayGame{
				Actions: []types.Action{
					{Description: "LeBron James dunk", TeamID: 10, Period: 1},
				},
			},
		}

		// Act
		updatedModel, _ := modelWithBoxScore.Update(PlayByPlayMsg(pbp))

		view := updatedModel.View()
		assert.Contains(t, view, "LeBron James dunk")
	})
}

func TestView_Layout(t *testing.T) {
	client := &mockNbaClient{}
	m := New(client, "123")
	m.boxScore = types.LiveBoxScoreResponse{
		Game: types.Game{
			GameId: "123",
		},
	}

	t.Run("horizontal layout", func(t *testing.T) {
		m.width = 100
		m.height = 40
		view := m.View()
		assert.Contains(t, view, "Box Scores")
		assert.Contains(t, view, "gamelog")
	})

	t.Run("vertical layout", func(t *testing.T) {
		m.width = 40
		m.height = 40
		view := m.View()
		assert.Contains(t, view, "Box Scores")
		assert.Contains(t, view, "gamelog")
	})

	t.Run("terminal too small", func(t *testing.T) {
		m.width = 10
		m.height = 5
		view := m.View()
		assert.Contains(t, view, "Terminal too small")
	})

	t.Run("available height too small", func(t *testing.T) {
		m.width = 50
		m.height = 10
		view := m.View()
		assert.Contains(t, view, "Terminal height too small")
	})
}

func TestUpdate_KeyEvents(t *testing.T) {
	client := &mockNbaClient{}
	m := New(client, "123")
	m.width = 100
	m.height = 40
	m.boxScore = types.LiveBoxScoreResponse{
		Game: types.Game{
			GameId:   "123",
			HomeTeam: types.Team{TeamId: 1, TeamTricode: "LAL", Players: &[]types.Player{{FamilyName: "James"}}},
			AwayTeam: types.Team{TeamId: 2, TeamTricode: "GSW"},
		},
	}

	t.Run("ctrl+q switch period", func(t *testing.T) {
		m.selectedPeriod = 1
		model, _ := m.Update(tea.KeyMsg{Type: tea.KeyCtrlQ})
		assert.Equal(t, 2, model.(Model).selectedPeriod)

		m.selectedPeriod = 4
		model, _ = m.Update(tea.KeyMsg{Type: tea.KeyCtrlQ})
		assert.Equal(t, 1, model.(Model).selectedPeriod)
	})

	t.Run("ctrl+b ctrl+l switch focus", func(t *testing.T) {
		model, _ := m.Update(tea.KeyMsg{Type: tea.KeyCtrlL})
		assert.Equal(t, gameLogFocus, model.(Model).focus)

		model, _ = model.Update(tea.KeyMsg{Type: tea.KeyCtrlB})
		assert.Equal(t, boxScoreFocus, model.(Model).focus)
	})

	t.Run("navigation down/up boxscore", func(t *testing.T) {
		m.focus = boxScoreFocus
		m.boxOffset = 0
		players := make([]types.Player, 5)
		m.boxScore.Game.HomeTeam.Players = &players

		model, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
		assert.Equal(t, 1, model.(Model).boxOffset)

		model, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("k")})
		assert.Equal(t, 0, model.(Model).boxOffset)
	})

	t.Run("navigation down/up gamelog", func(t *testing.T) {
		m.focus = gameLogFocus
		m.logOffset = 0
		m.selectedPeriod = 1
		m.pbp.Game.Actions = []types.Action{
			{Period: 1, TeamID: 1, Description: "A1"},
			{Period: 1, TeamID: 1, Description: "A2"},
		}

		model, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
		assert.Equal(t, 1, model.(Model).logOffset)

		model, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("k")})
		assert.Equal(t, 0, model.(Model).logOffset)
	})

	t.Run("ctrl+c quit", func(t *testing.T) {
		_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyCtrlC})
		assert.NotNil(t, cmd)
	})
}

func TestModel_FetchFunctions(t *testing.T) {
	client := &mockNbaClient{}
	m := New(client, "123")

	t.Run("fetch box score success", func(t *testing.T) {
		msg := m.fetchBoxScore()
		_, ok := msg.(BoxScoreMsg)
		assert.True(t, ok)
	})

	t.Run("fetch playbyplay success", func(t *testing.T) {
		msg := m.fetchPlayByPlay()
		_, ok := msg.(PlayByPlayMsg)
		assert.True(t, ok)
	})

	t.Run("fetch failure", func(t *testing.T) {
		client.err = fmt.Errorf("api error")
		msg := m.fetchBoxScore()
		assert.Equal(t, client.err, msg.(ErrorMsg))

		msg = m.fetchPlayByPlay()
		assert.Equal(t, client.err, msg.(ErrorMsg))
		client.err = nil
	})
}

func TestView_RenderEdgeCases(t *testing.T) {
	client := &mockNbaClient{}
	m := New(client, "123")
	m.width = 100
	m.height = 40

	t.Run("player without statistics", func(t *testing.T) {
		players := []types.Player{{FamilyName: "James", Statistics: nil}}
		m.boxScore.Game.GameId = "123"
		m.boxScore.Game.HomeTeam.Players = &players
		view := m.View()
		assert.Contains(t, view, "James")
	})

	t.Run("gamelog truncation", func(t *testing.T) {
		m.width = 30
		m.pbp.Game.Actions = []types.Action{
			{Period: 1, TeamID: 1, Clock: "12:00", Description: "Very long description that should be truncated"},
		}
		m.boxScore.Game.HomeTeam.TeamId = 1
		view := m.View()
		assert.Contains(t, view, "...")
	})
}
func TestUpdate_SwitchTeam(t *testing.T) {
	client := &mockNbaClient{}
	m := New(client, "123")
	m.boxScore = types.LiveBoxScoreResponse{
		Game: types.Game{
			GameId:   "123",
			HomeTeam: types.Team{TeamTricode: "LAL"},
			AwayTeam: types.Team{TeamTricode: "GSW"},
		},
	}
	m.width = 100
	m.height = 40

	// Initial should show LAL and GSW in header
	view := m.View()
	if !strings.Contains(view, "LAL") || !strings.Contains(view, "GSW") {
		t.Errorf("Initial view should show LAL and GSW in header")
	}

	// Switch to Away
	updatedModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyCtrlS})
	if updatedModel.(Model).IsShowingHome() {
		t.Errorf("Should be showing away after Ctrl+S")
	}

	// Switch back to Home
	updatedModel, _ = updatedModel.Update(tea.KeyMsg{Type: tea.KeyCtrlS})
	if !updatedModel.(Model).IsShowingHome() {
		t.Errorf("Should be showing home after second Ctrl+S")
	}
}

func TestView_BoxScoreTable(t *testing.T) {
	client := &mockNbaClient{}
	m := New(client, "123")
	m.width = 100
	m.height = 40

	players := []types.Player{
		{
			FirstName:  "LeBron",
			FamilyName: "James",
			Statistics: &types.PlayerBoxScoreStatistic{
				CommonBoxScoreStatistic: types.CommonBoxScoreStatistic{
					Minutes: "35:00",
					Pts:     ptr(30),
					Reb:     ptr(10),
					Ast:     ptr(10),
				},
			},
		},
	}

	m.boxScore = types.LiveBoxScoreResponse{
		Game: types.Game{
			GameId: "123",
			HomeTeam: types.Team{
				TeamTricode: "LAL",
				Players:     &players,
			},
		},
	}

	view := m.View()
	headers := []string{"PLAYER", "MIN", "PTS", "REB", "AST"}
	for _, h := range headers {
		if !strings.Contains(view, h) {
			t.Errorf("View should contain table header %s", h)
		}
	}

	if !strings.Contains(view, "L.James") {
		t.Errorf("View should contain player name L.James, got: %s", view)
	}
}

func TestView_BoldWinner(t *testing.T) {
	lipgloss.SetColorProfile(termenv.ANSI256)
	client := &mockNbaClient{}
	m := New(client, "123")
	m.width = 100
	m.height = 40
	m.boxScore = types.LiveBoxScoreResponse{
		Game: types.Game{
			GameId:   "123",
			HomeTeam: types.Team{TeamTricode: "LAL", Score: 110},
			AwayTeam: types.Team{TeamTricode: "GSW", Score: 100},
		},
	}

	view := m.View()
	// \x1b[1m is bold, \x1b[0m is reset
	if !strings.Contains(view, "\x1b[1mLAL\x1b[0m") {
		t.Errorf("Winning team LAL should be bolded, got: %q", view)
	}
}

func TestView_Footer(t *testing.T) {
	client := &mockNbaClient{}
	m := New(client, "123")
	m.width = 100
	m.height = 40
	m.boxScore = types.LiveBoxScoreResponse{
		Game: types.Game{GameId: "123"},
	}
	fixedTime := time.Date(2023, 10, 27, 10, 0, 0, 0, time.Local)
	m.SetLastUpdated(fixedTime)

	view := m.View()
	if !strings.Contains(view, "Last updated:") {
		t.Errorf("View should contain 'Last updated:', got: %s", view)
	}
	if !strings.Contains(view, "<hjkli←↓↑→ >: move") {
		t.Errorf("View should contain help text, got: %s", view)
	}
}

func TestUpdate_Navigation(t *testing.T) {
	client := &mockNbaClient{}
	m := New(client, "123")
	m.width = 100
	m.height = 40
	m.boxScore = types.LiveBoxScoreResponse{
		Game: types.Game{
			GameId:   "123",
			HomeTeam: types.Team{TeamTricode: "LAL"},
			AwayTeam: types.Team{TeamTricode: "GSW"},
		},
	}
	m.pbp = types.LivePlayByPlayResponse{
		Game: types.PlayByPlayGame{
			Actions: []types.Action{{Description: "A1"}, {Description: "A2"}},
		},
	}

	// Initial focus should be BoxScore (0)
	if m.GetFocus() != 0 {
		t.Errorf("Initial focus should be 0")
	}

	// Move right to GameLog using Ctrl+L
	m2, _ := m.Update(tea.KeyMsg{Type: tea.KeyCtrlL})
	if m2.(Model).GetFocus() != 1 {
		t.Errorf("Focus should be 1 after Ctrl+L")
	}

	// Move back to BoxScore using Ctrl+B
	m3, _ := m2.Update(tea.KeyMsg{Type: tea.KeyCtrlB})
	if m3.(Model).GetFocus() != 0 {
		t.Errorf("Focus should be 0 after Ctrl+B")
	}

	// Check Selected Team display
	view := m3.View()
	if !strings.Contains(view, "Selected Team: LAL") {
		t.Errorf("View should show Selected Team: LAL, got: %s", view)
	}

	// Switch team and check display
	m4, _ := m3.Update(tea.KeyMsg{Type: tea.KeyCtrlS})
	if !strings.Contains(m4.View(), "Selected Team: GSW") {
		t.Errorf("View should show Selected Team: GSW after switch")
	}
}

func TestView_GameLogFiltering(t *testing.T) {
	lipgloss.SetColorProfile(termenv.ANSI256)
	client := &mockNbaClient{}
	m := New(client, "123")
	m.width = 100
	m.height = 40
	m.boxScore = types.LiveBoxScoreResponse{
		Game: types.Game{
			GameId:   "123",
			HomeTeam: types.Team{TeamTricode: "LAL", TeamId: 1},
			AwayTeam: types.Team{TeamTricode: "GSW", TeamId: 2},
		},
	}
	m.pbp = types.LivePlayByPlayResponse{
		Game: types.PlayByPlayGame{
			Actions: []types.Action{
				{Description: "LAL Action Q1", TeamID: 1, Period: 1},
				{Description: "GSW Action Q1", TeamID: 2, Period: 1},
				{Description: "GSW Action Q2", TeamID: 2, Period: 2},
			},
		},
	}

	// 1. Initial State: Team LAL, Period 1
	view1 := m.View()
	assert.Contains(t, view1, "LAL Action Q1")
	assert.NotContains(t, view1, "GSW Action Q1")

	// 2. Switch to Team GSW (Ctrl+S) -> Still Period 1
	m2, _ := m.Update(tea.KeyMsg{Type: tea.KeyCtrlS})
	view2 := m2.View()
	assert.Contains(t, view2, "GSW Action Q1")
	assert.NotContains(t, view2, "LAL Action Q1")

	// 3. Switch to Period 2 (Ctrl+Q) -> Team GSW
	m3, _ := m2.Update(tea.KeyMsg{Type: tea.KeyCtrlQ})
	view3 := m3.View()
	assert.Contains(t, view3, "GSW Action Q2")
	assert.NotContains(t, view3, "GSW Action Q1")

	// 4. Check period selector UI
	assert.Contains(t, view3, "1")
	assert.Contains(t, view3, "Q")
	assert.Contains(t, view3, "2")
}

func TestModel_GettersAndSetters(t *testing.T) {
	client := &mockNbaClient{}
	m := New(client, "123")
	
	m.showingHome = false
	assert.False(t, m.IsShowingHome())
	
	m.focus = gameLogFocus
	assert.Equal(t, int(gameLogFocus), m.GetFocus())
	
	m.logOffset = 5
	assert.Equal(t, 5, m.GetLogOffset())
	
	m.boxOffset = 10
	assert.Equal(t, 10, m.GetBoxOffset())
	
	m.selectedPeriod = 3
	assert.Equal(t, 3, m.GetSelectedPeriod())
	
	now := time.Now()
	m.SetLastUpdated(now)
	assert.Equal(t, now, m.lastUpdated)
}

func TestUpdate_KeyEvents_Boundaries(t *testing.T) {
	client := &mockNbaClient{}
	m := New(client, "123")
	m.boxScore.Game.HomeTeam.Players = &[]types.Player{{FamilyName: "P1"}}
	
	t.Run("boxscore boundaries", func(t *testing.T) {
		m.focus = boxScoreFocus
		m.boxOffset = 0
		// Up at 0
		model, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("k")})
		assert.Equal(t, 0, model.(Model).boxOffset)
		
		// Down at end (only 1 player)
		model, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
		assert.Equal(t, 0, model.(Model).boxOffset)
	})
	
	t.Run("gamelog boundaries", func(t *testing.T) {
		m.focus = gameLogFocus
		m.logOffset = 0
		m.pbp.Game.Actions = []types.Action{{Period: 1, TeamID: 0, Description: "A1"}}
		m.boxScore.Game.HomeTeam.TeamId = 0
		m.selectedPeriod = 1
		
		// Up at 0
		model, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("k")})
		assert.Equal(t, 0, model.(Model).logOffset)
		
		// Down at end
		model, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
		assert.Equal(t, 0, model.(Model).logOffset)
	})
}
