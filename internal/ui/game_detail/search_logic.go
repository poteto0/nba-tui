package game_detail

import (
	"strings"

	"github.com/poteto0/go-nba-sdk/types"
)

func SearchActions(actions []types.Action, query string) []int {
	if query == "" {
		return []int{}
	}

	indices := []int{}
	queryLower := strings.ToLower(query)

	for i, action := range actions {
		if strings.Contains(strings.ToLower(action.Description), queryLower) {
			indices = append(indices, i)
		}
	}
	return indices
}
