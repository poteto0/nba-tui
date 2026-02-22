package game_detail

import (
	"testing"

	"github.com/poteto0/go-nba-sdk/types"
	"github.com/stretchr/testify/assert"
)

func TestGetKawaiiPrefix(t *testing.T) {
	tests := []struct {
		name     string
		stats    types.PlayerBoxScoreStatistic
		expected string
	}{
		{
			name: "Triple Double",
			stats: types.PlayerBoxScoreStatistic{
				CommonBoxScoreStatistic: types.CommonBoxScoreStatistic{
					Pts: new(10), Reb: new(10), Ast: new(10), Stl: new(0), Blk: new(0),
				},
			},
			expected: "👑",
		},
		{
			name: "5x5",
			stats: types.PlayerBoxScoreStatistic{
				CommonBoxScoreStatistic: types.CommonBoxScoreStatistic{
					Pts: new(5), Reb: new(5), Ast: new(5), Stl: new(5), Blk: new(5),
				},
			},
			expected: "💯",
		},
		{
			name: "Sniper (3PM>=8, 3P%>=50)",
			stats: types.PlayerBoxScoreStatistic{
				CommonBoxScoreStatistic: types.CommonBoxScoreStatistic{
					Fg3M: new(8), Fg3Pct: new(0.50),
				},
			},
			expected: "🎯",
		},
		{
			name: "Alien (PTS>=50)",
			stats: types.PlayerBoxScoreStatistic{
				CommonBoxScoreStatistic: types.CommonBoxScoreStatistic{
					Pts: new(50),
				},
			},
			expected: "👽",
		},
		{
			name: "Strong (REB>=20)",
			stats: types.PlayerBoxScoreStatistic{
				CommonBoxScoreStatistic: types.CommonBoxScoreStatistic{
					Reb: new(20),
				},
			},
			expected: "💪",
		},
		{
			name: "TeamPlayer (AST>=20)",
			stats: types.PlayerBoxScoreStatistic{
				CommonBoxScoreStatistic: types.CommonBoxScoreStatistic{
					Ast: new(20),
				},
			},
			expected: "🤝",
		},
		{
			name: "Guardian (BLK>=7)",
			stats: types.PlayerBoxScoreStatistic{
				CommonBoxScoreStatistic: types.CommonBoxScoreStatistic{
					Blk: new(7),
				},
			},
			expected: "🔒",
		},
		{
			name: "Thief (STL>=5)",
			stats: types.PlayerBoxScoreStatistic{
				CommonBoxScoreStatistic: types.CommonBoxScoreStatistic{
					Stl: new(5),
				},
			},
			expected: "🥷",
		},
		{
			name: "Multiple (Triple Double + 5x5)",
			stats: types.PlayerBoxScoreStatistic{
				CommonBoxScoreStatistic: types.CommonBoxScoreStatistic{
					Pts: new(10), Reb: new(10), Ast: new(10), Stl: new(10), Blk: new(10),
				},
			},
			expected: "👑",
		},
		{
			name: "No Match",
			stats: types.PlayerBoxScoreStatistic{
				CommonBoxScoreStatistic: types.CommonBoxScoreStatistic{
					Pts: new(0),
				},
			},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetKawaiiPrefix(tt.stats)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestShouldUnderlineStat(t *testing.T) {
	tests := []struct {
		name     string
		statName string
		val      *int
		expected bool
	}{
		{"PTS >= 10", "PTS", new(10), true},
		{"PTS < 10", "PTS", new(9), false},
		{"REB >= 10", "REB", new(10), true},
		{"AST >= 10", "AST", new(10), true},
		{"STL > 3", "STL", new(4), true},
		{"STL <= 3", "STL", new(3), false},
		{"BLK > 3", "BLK", new(4), true},
		{"BLK <= 3", "BLK", new(3), false},
		{"Other", "FGM", new(10), false},
		{"Nil", "PTS", nil, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ShouldUnderlineStat(tt.statName, tt.val)
			assert.Equal(t, tt.expected, result)
		})
	}
}
