package game_detail

import (
	"github.com/poteto0/go-nba-sdk/types"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestBoxScoreTotalRow(t *testing.T) {
	client := &mockNbaClient{}
	m := New(client, "123")
	m.width = 200
	m.height = 40

	pts := 10
	reb := 5
	ast := 2

	// Create a team with statistics
	teamStats := types.TeamBoxScoreStatistic{
		CommonBoxScoreStatistic: types.CommonBoxScoreStatistic{
			Pts: &pts,
			Reb: &reb,
			Ast: &ast,
		},
	}

	m.boxScore = types.LiveBoxScoreResponse{
		Game: types.Game{
			GameId: "123",
			HomeTeam: types.Team{
				TeamTricode: "LAL",
				Statistics:  &teamStats,
				Players:     &[]types.Player{},
			},
		},
	}

	view := m.View()

	assert.Contains(t, view, "TOTAL", "Box score should contain 'TOTAL' row")
	assert.Contains(t, view, "10", "Box score total should contain points")
}
