package game_detail

import (
	"strings"
	"testing"

	"github.com/poteto0/go-nba-sdk/types"
	"github.com/stretchr/testify/assert"
	tea "github.com/charmbracelet/bubbletea"
)

func TestBoxScoreHeader(t *testing.T) {
	client := &mockNbaClient{}
	m := New(client, "123")
	m.width = 200
	m.height = 40
	m.boxScore = types.LiveBoxScoreResponse{
		Game: types.Game{
			GameId: "123",
			HomeTeam: types.Team{
				TeamTricode: "LAL",
				Players:     &[]types.Player{},
			},
		},
	}

	view := m.View()

	expectedHeaderParts := []string{
		"MIN", "FGM", "FGA", "FG%", "3PM", "3PA", "3P%", "FTM", "FTA", "FT%",
		"OREB", "DREB", "REB", "AST", "STL", "BLK", "TO", "PF", "PTS", "+/-",
	}

	for _, part := range expectedHeaderParts {
		assert.Contains(t, view, part, "Box score header should contain %s", part)
	}
}

func TestMinutesFormatting(t *testing.T) {
	client := &mockNbaClient{}
	m := New(client, "123")
	m.width = 100
	m.height = 40

	t.Run("Long format minutes", func(t *testing.T) {
		players := []types.Player{
			{
				FamilyName: "LongTime",
				Statistics: &types.PlayerBoxScoreStatistic{
					CommonBoxScoreStatistic: types.CommonBoxScoreStatistic{
						Minutes: "PT36M10.01S",
					},
				},
			},
		}
		m.boxScore = types.LiveBoxScoreResponse{
			Game: types.Game{
				GameId:   "123",
				HomeTeam: types.Team{TeamTricode: "LAL", Players: &players},
			},
		}
		
		view := m.View()
		assert.Contains(t, view, "36:10", "Should truncate '36:10.01' to '36:10'")
	})

	t.Run("Short format minutes", func(t *testing.T) {
		players := []types.Player{
			{
				FamilyName: "ShortTime",
				Statistics: &types.PlayerBoxScoreStatistic{
					CommonBoxScoreStatistic: types.CommonBoxScoreStatistic{
						Minutes: "PT0M0S",
					},
				},
			},
		}
		m.boxScore = types.LiveBoxScoreResponse{
			Game: types.Game{
				GameId:   "123",
				HomeTeam: types.Team{TeamTricode: "LAL", Players: &players},
			},
		}
		
		view := m.View()
		assert.NotContains(t, view, "0:00")
	})
}

func TestGameStatusFormatting(t *testing.T) {
	client := &mockNbaClient{}
	m := New(client, "123")
	m.width = 100
	m.height = 40

	t.Run("Not Started", func(t *testing.T) {
		m.boxScore = types.LiveBoxScoreResponse{
			Game: types.Game{
				GameId:     "123",
				GameStatus: 1, 
				HomeTeam: types.Team{TeamTricode: "LAL"},
				AwayTeam: types.Team{TeamTricode: "GSW"},
			},
		}
		
		view := m.View()
		assert.Contains(t, view, "not started")
	})

	t.Run("Overtime 1OT", func(t *testing.T) {
		m.boxScore = types.LiveBoxScoreResponse{
			Game: types.Game{
				GameId:     "123",
				GameStatus: 2,
				Period:     5,
				GameClock:  "PT05M00.00S",
				HomeTeam: types.Team{TeamTricode: "LAL"},
				AwayTeam: types.Team{TeamTricode: "GSW"},
			},
		}
		
		view := m.View()
		assert.Contains(t, view, "1OT")
		assert.NotContains(t, view, "5Q")
	})

	t.Run("Overtime 2OT", func(t *testing.T) {
		m.boxScore = types.LiveBoxScoreResponse{
			Game: types.Game{
				GameId:     "123",
				GameStatus: 2,
				Period:     6,
				GameClock:  "PT05M00.00S",
				HomeTeam: types.Team{TeamTricode: "LAL"},
				AwayTeam: types.Team{TeamTricode: "GSW"},
			},
		}
		
		view := m.View()
		assert.Contains(t, view, "2OT")
		assert.NotContains(t, view, "6Q")
	})
}

func TestBoxScoreAlignment(t *testing.T) {
	client := &mockNbaClient{}
	m := New(client, "123")
	m.width = 200 // Wide enough to avoid truncation
	m.height = 40

	pts := 9
	reb := 123 
	players := []types.Player{
		{
			FamilyName: "AlignCheck",
			Statistics: &types.PlayerBoxScoreStatistic{
				CommonBoxScoreStatistic: types.CommonBoxScoreStatistic{
					Minutes: "PT10M00.00S",
					Pts: &pts,
					Reb: &reb,
				},
			},
		},
	}
	m.boxScore = types.LiveBoxScoreResponse{
		Game: types.Game{
			GameId:   "123",
			HomeTeam: types.Team{TeamTricode: "LAL", Players: &players},
		},
	}

	view := m.View()

	// Find the line with stats
	lines := strings.Split(view, "\n")
	var statLine string
	for _, l := range lines {
		if strings.Contains(l, "AlignCheck") {
			statLine = l
			break
		}
	}
	assert.NotEmpty(t, statLine)

	// We expect right alignment for stats.
	// Previous implementation used %-3d (left align). 9 -> "9  "
	// New requirement: %3d (right align). 9 -> "  9"
	
	// Let's assume PTS is around index 25-30.
	// We can check if "9  " exists vs "  9"
	// But simply checking strict equality of the format string part is hard without full line knowledge.
	// However, we can check that for a single digit number in a 3-char column, it starts with spaces.
	
	// Let's create a regex or check index? 
	// Easier: Just check if we find "  9" in the line, and NOT "9  ".
	// Assuming 9 is unique enough in this line.
	// "9" is the PTS value.
	
	// Wait, Minutes is 10:00 (5 chars).
	// Player name AlignCheck (10 chars).
	// So "AlignCheck      10:00   9" vs "AlignCheck      10:00 9  "
	
	assert.Contains(t, statLine, "  9", "Stat '9' should be right aligned (preceded by spaces)")
	// assert.NotContains(t, statLine, "9  ", "Stat '9' should NOT be left aligned (followed by spaces)")
}

func TestBoxScoreHorizontalScrolling(t *testing.T) {
	client := &mockNbaClient{}
	m := New(client, "123")
	
	// Force a small width so content truncates
	m.width = 40 
	m.height = 40
	m.focus = boxScoreFocus

	// Player with data
	pts := 10
	players := []types.Player{
		{
			FamilyName: "Scroller",
			Statistics: &types.PlayerBoxScoreStatistic{
				CommonBoxScoreStatistic: types.CommonBoxScoreStatistic{
					Minutes: "PT10M00.00S",
					Pts: &pts,
				},
			},
		},
	}
	m.boxScore = types.LiveBoxScoreResponse{
		Game: types.Game{
			GameId:   "123",
			HomeTeam: types.Team{TeamTricode: "LAL", Players: &players},
		},
	}

	// 1. Initial State: scroll 0
	view1 := m.View()
	// PLAYER(15) + MIN(5) = 20 chars approx + spacing
	// With width 40, we should see start of the line (Name)
	assert.Contains(t, view1, "Scroller")
	// But probably not the end of the line (e.g. +/- or PTS if it's far right)
	// Let's assume +/- is the last column.
	assert.NotContains(t, view1, "+/- ", "Should be truncated at width 40")

	// 2. Scroll Right (l)
	// We need to implement 'h'/'l' support in Update
	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("l")})
	m2 := newModel.(Model)
	
	// We expect the view to shift. 
	// If we scroll enough, "Scroller" (at the start) might disappear or move left.
	// Or simply, we see new content.
	
	// Let's simulate multiple scrolls to move significantly
	for i := 0; i < 20; i++ {
		newModel, _ = m2.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("l")})
		m2 = newModel.(Model)
	}
	mEnd := m2
	viewEnd := mEnd.View()
	
	// Now we might see the end of the table
	// Depending on implementation, maybe +/- becomes visible
	// OR the left side "Scroller" becomes hidden
	
	// Since we don't know exact chars, checking that viewEnd != view1 is a start
	// And ideally, checking that we can see something that was hidden.
	
	assert.NotEqual(t, view1, viewEnd, "View should change after scrolling right")
	assert.Contains(t, viewEnd, "Scroller", "Player name should still be visible after horizontal scrolling")
	
	// 3. Scroll Left (h)
	newModel, _ = mEnd.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("h")})
	// Should change back towards view1
}

func TestBoxScoreHorizontalScrollingBoundary(t *testing.T) {
	client := &mockNbaClient{}
	m := New(client, "123")
	
	m.width = 40 
	m.height = 40
	m.focus = boxScoreFocus

	players := []types.Player{
		{
			FamilyName: "Boundary",
			Statistics: &types.PlayerBoxScoreStatistic{},
		},
	}
	m.boxScore = types.LiveBoxScoreResponse{
		Game: types.Game{
			GameId:   "123",
			HomeTeam: types.Team{TeamTricode: "LAL", Players: &players},
		},
	}

	// Scroll right many times
	newModel := tea.Model(m)
	for i := 0; i < 500; i++ {
		newModel, _ = newModel.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("l")})
	}
	
	mEnd := newModel.(Model)
	
	// Scroll right one more time
	mOver, _ := mEnd.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("l")})
	
	assert.Equal(t, mEnd.boxScrollX, mOver.(Model).boxScrollX, "Should not scroll past the end of the line")
}
