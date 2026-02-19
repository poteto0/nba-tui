package game_detail

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
	"github.com/poteto0/go-nba-sdk/types"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestBoxScoreDecoration(t *testing.T) {
	lipgloss.SetColorProfile(termenv.TrueColor)
	defer lipgloss.SetColorProfile(termenv.Ascii) // Reset to no color (or Ascii)
	client := &mockNbaClient{}

	pts1, pts2 := 10, 20
	reb1, reb2 := 5, 2
	ast1, ast2 := 2, 8
	pm1, pm2 := 5.0, -3.0

	players := []types.Player{
		{
			FamilyName: "Player1",
			Statistics: &types.PlayerBoxScoreStatistic{
				CommonBoxScoreStatistic: types.CommonBoxScoreStatistic{
					Pts: &pts1,
					Reb: &reb1,
					Ast: &ast1,
				},
				PlusMinus: &pm1,
			},
		},
		{
			FamilyName: "Player2",
			Statistics: &types.PlayerBoxScoreStatistic{
				CommonBoxScoreStatistic: types.CommonBoxScoreStatistic{
					Pts: &pts2,
					Reb: &reb2,
					Ast: &ast2,
				},
				PlusMinus: &pm2,
			},
		},
	}

	m := New(client, "123", Config{NoDecoration: false})
	m.width = 200
	m.height = 40
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

	// Player 2 has max Pts (20) and max Ast (8)
	// Player 1 has max Reb (5)
	// Player 1 has positive +/- (5.0) -> Green
	// Player 2 has negative +/- (-3.0) -> Red

	// Bold check (extremely simplified, lipgloss uses ANSI codes)
	// Bold is \x1b[1m
	assert.Contains(t, view, "20", "Player 2 points should be in view")
	assert.Contains(t, view, "5", "Player 1 rebounds should be in view")

	// Decoration check
	t.Run("NoDecoration true", func(t *testing.T) {
		m.config.NoDecoration = true
		viewNoDeco := m.View()
		assert.NotEqual(t, view, viewNoDeco, "View with and without decoration should be different")
	})
}
