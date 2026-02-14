package game_detail

import (
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
	"github.com/poteto0/go-nba-sdk/types"
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
	var _ tea.Model = m
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

func TestUpdate_BoxScore(t *testing.T) {
	client := &mockNbaClient{}
	m := New(client, "123")
	m.width = 100
	m.height = 40

	// Create dummy data
	boxScore := types.LiveBoxScoreResponse{
		Game: types.Game{
			GameId:   "123",
			HomeTeam: types.Team{TeamName: "Lakers", TeamTricode: "LAL"},
			AwayTeam: types.Team{TeamName: "Warriors", TeamTricode: "GSW"},
		},
	}

	updatedModel, _ := m.Update(BoxScoreMsg(boxScore))

	view := updatedModel.View()
	if !strings.Contains(view, "LAL") || !strings.Contains(view, "GSW") {
		t.Errorf("View should contain team tricodes after BoxScoreMsg, got: %s", view)
	}
}

func TestUpdate_PlayByPlay(t *testing.T) {
	client := &mockNbaClient{}
	m := New(client, "123")
	m.width = 100
	m.height = 40

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

	updatedModel, _ := modelWithBoxScore.Update(PlayByPlayMsg(pbp))

	view := updatedModel.View()
	if !strings.Contains(view, "LeBron James dunk") {
		t.Errorf("View should contain pbp description, got: %s", view)
	}
}

func TestView_Layout(t *testing.T) {
	client := &mockNbaClient{}
	m := New(client, "123")
	m.boxScore = types.LiveBoxScoreResponse{
		Game: types.Game{
			GameId: "123",
		},
	}
	// Set size to avoid "Too small"
	m.width = 100
	m.height = 40
	
	view := m.View()
	if !strings.Contains(view, "Box Scores") || !strings.Contains(view, "gamelog") {
		t.Errorf("View should contain layout headers, got: %s", view)
	}
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
			FirstName: "LeBron",
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
	if !strings.Contains(view1, "LAL Action Q1") {
		t.Errorf("Should contain LAL Action Q1")
	}
	if strings.Contains(view1, "GSW Action Q1") {
		t.Errorf("Should NOT contain GSW Action Q1 (filtered by team)")
	}

	// 2. Switch to Team GSW (Ctrl+S) -> Still Period 1
	m2, _ := m.Update(tea.KeyMsg{Type: tea.KeyCtrlS})
	view2 := m2.View()
	if !strings.Contains(view2, "GSW Action Q1") {
		t.Errorf("Should contain GSW Action Q1 after switching team")
	}
	if strings.Contains(view2, "LAL Action Q1") {
		t.Errorf("Should NOT contain LAL Action Q1 after switching team")
	}

	// 3. Switch to Period 2 (Ctrl+Q) -> Team GSW
	m3, _ := m2.Update(tea.KeyMsg{Type: tea.KeyCtrlQ})
	view3 := m3.View()
	if !strings.Contains(view3, "GSW Action Q2") {
		t.Errorf("Should contain GSW Action Q2 after switching period")
	}
	if strings.Contains(view3, "GSW Action Q1") {
		t.Errorf("Should NOT contain GSW Action Q1 (filtered by period)")
	}
	
	// 4. Check period selector UI
	// Since 2Q might be styled like \x1b[4m2\x1b[0m\x1b[4mQ\x1b[0m, we check separately
	if !strings.Contains(view3, "1") || !strings.Contains(view3, "Q") || !strings.Contains(view3, "2") {
		t.Errorf("Should show period selector elements, got: %q", view3)
	}
}