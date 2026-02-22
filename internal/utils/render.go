package utils

import (
	"fmt"

	"github.com/poteto0/go-nba-sdk/types"
)

func RenderGameStatus(game types.Game) string {
	switch {
	case !game.IsGameStart():
		return "Not Started"
	case game.IsFinished():
		return "Final"
	default:
		periodStr := fmt.Sprintf("%dQ", game.Period)
		if game.IsOverTime() {
			periodStr = fmt.Sprintf("%dOT", game.OverTimeNum())
		}

		clock := game.Clock()
		if len(clock) > 5 {
			clock = clock[:5]
		} else {
			clock = "-"
		}

		return fmt.Sprintf("%s (%s)", periodStr, clock)
	}
}
