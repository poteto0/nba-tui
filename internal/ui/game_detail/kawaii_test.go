package game_detail

import (
	"testing"

	"github.com/poteto0/go-nba-sdk/types"
	"github.com/stretchr/testify/assert"
)

func TestGetKawaiiPrefix(t *testing.T) {
	ptrInt := func(i int) *int { return &i }
	ptrFloat := func(f float64) *float64 { return &f }

	tests := []struct {
		name     string
		stats    types.PlayerBoxScoreStatistic
		expected string
	}{
		{
			name: "Triple Double",
			stats: types.PlayerBoxScoreStatistic{
				CommonBoxScoreStatistic: types.CommonBoxScoreStatistic{
					Pts: ptrInt(10), Reb: ptrInt(10), Ast: ptrInt(10), Stl: ptrInt(0), Blk: ptrInt(0),
				},
			},
			expected: "👑",
		},
		{
			name: "5x5",
			stats: types.PlayerBoxScoreStatistic{
				CommonBoxScoreStatistic: types.CommonBoxScoreStatistic{
					Pts: ptrInt(5), Reb: ptrInt(5), Ast: ptrInt(5), Stl: ptrInt(5), Blk: ptrInt(5),
				},
			},
			expected: "💯",
		},
		{
			name: "Sniper (3PM>=8, 3P%>=50)",
			stats: types.PlayerBoxScoreStatistic{
				CommonBoxScoreStatistic: types.CommonBoxScoreStatistic{
					Fg3M: ptrInt(8), Fg3Pct: ptrFloat(0.50),
				},
			},
			expected: "🎯",
		},
		{
			name: "Alien (PTS>=50)",
			stats: types.PlayerBoxScoreStatistic{
				CommonBoxScoreStatistic: types.CommonBoxScoreStatistic{
					Pts: ptrInt(50),
				},
			},
			expected: "👽",
		},
		{
			name: "Strong (REB>=20)",
			stats: types.PlayerBoxScoreStatistic{
				CommonBoxScoreStatistic: types.CommonBoxScoreStatistic{
					Reb: ptrInt(20),
				},
			},
			expected: "💪",
		},
		{
			name: "TeamPlayer (AST>=20)",
			stats: types.PlayerBoxScoreStatistic{
				CommonBoxScoreStatistic: types.CommonBoxScoreStatistic{
					Ast: ptrInt(20),
				},
			},
			expected: "🤝",
		},
		{
			name: "Guardian (BLK>=7)",
			stats: types.PlayerBoxScoreStatistic{
				CommonBoxScoreStatistic: types.CommonBoxScoreStatistic{
					Blk: ptrInt(7),
				},
			},
			expected: "🛡️",
		},
		{
			name: "Thief (STL>=5)",
			stats: types.PlayerBoxScoreStatistic{
				CommonBoxScoreStatistic: types.CommonBoxScoreStatistic{
					Stl: ptrInt(5),
				},
			},
			expected: "🥷🏻",
		},
		{
			name: "Multiple (Triple Double + 5x5)",
			stats: types.PlayerBoxScoreStatistic{
				CommonBoxScoreStatistic: types.CommonBoxScoreStatistic{
					Pts: ptrInt(10), Reb: ptrInt(10), Ast: ptrInt(10), Stl: ptrInt(10), Blk: ptrInt(10),
				},
			},
			expected: "👑",
		},
		{
			name: "No Match",
			stats: types.PlayerBoxScoreStatistic{
				CommonBoxScoreStatistic: types.CommonBoxScoreStatistic{
					Pts: ptrInt(0),
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
	ptrInt := func(i int) *int { return &i }

	tests := []struct {
		name     string
		statName string
		val      *int
		expected bool
	}{
		{"PTS >= 10", "PTS", ptrInt(10), true},
		{"PTS < 10", "PTS", ptrInt(9), false},
		{"REB >= 10", "REB", ptrInt(10), true},
		{"AST >= 10", "AST", ptrInt(10), true},
		{"STL > 3", "STL", ptrInt(4), true},
		{"STL <= 3", "STL", ptrInt(3), false},
		{"BLK > 3", "BLK", ptrInt(4), true},
		{"BLK <= 3", "BLK", ptrInt(3), false},
		{"Other", "FGM", ptrInt(10), false},
		{"Nil", "PTS", nil, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ShouldUnderlineStat(tt.statName, tt.val)
			assert.Equal(t, tt.expected, result)
		})
	}
}
