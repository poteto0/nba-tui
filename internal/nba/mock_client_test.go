package nba

import (
	"testing"
	"github.com/stretchr/testify/assert"
	"nba-tui/internal/ui/root"
)

func TestMockClient_Interface(t *testing.T) {
	var _ root.Client = (*MockClient)(nil)
}

func TestMockClient_GetScoreboard(t *testing.T) {
	client := NewMockClient()
	games, err := client.GetScoreboard()
	
	assert.NoError(t, err)
	assert.Len(t, games, 2)
	assert.Equal(t, "LAL", games[0].HomeTeam.TeamTricode)
	assert.Equal(t, "GSW", games[0].AwayTeam.TeamTricode)
}

func TestMockClient_GetBoxScore(t *testing.T) {
	client := NewMockClient()
	res, err := client.GetBoxScore("0012300001")
	
	assert.NoError(t, err)
	assert.Equal(t, "0012300001", res.Game.GameId)
	assert.NotNil(t, res.Game.HomeTeam.Players)
	assert.True(t, len(*res.Game.HomeTeam.Players) > 0)
}

func TestMockClient_GetPlayByPlay(t *testing.T) {
	client := NewMockClient()
	res, err := client.GetPlayByPlay("0012300001")
	
	assert.NoError(t, err)
	assert.True(t, len(res.Game.Actions) > 0)
}
