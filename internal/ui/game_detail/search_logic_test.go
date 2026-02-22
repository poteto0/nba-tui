package game_detail

import (
	"testing"

	"github.com/poteto0/go-nba-sdk/types"
	"github.com/stretchr/testify/assert"
)

func TestSearchActions(t *testing.T) {
	actions := []types.Action{
		{Description: "Shai Gilgeous-Alexander makes 2-pt driving layup"},
		{Description: "Luguentz Dort misses 3-pt jump shot"},
		{Description: "Shai Gilgeous-Alexander defensive rebound"},
		{Description: "Jalen Williams makes free throw 1 of 2"},
	}

	tests := []struct {
		name     string
		query    string
		expected []int
	}{
		{
			name:     "match single",
			query:    "Dort",
			expected: []int{1},
		},
		{
			name:     "match multiple",
			query:    "Shai",
			expected: []int{0, 2},
		},
		{
			name:     "case insensitive",
			query:    "shai",
			expected: []int{0, 2},
		},
		{
			name:     "no match",
			query:    "Curry",
			expected: []int{},
		},
		{
			name:     "empty query",
			query:    "",
			expected: []int{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SearchActions(actions, tt.query)
			assert.Equal(t, tt.expected, result)
		})
	}
}
