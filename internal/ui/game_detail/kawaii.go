package game_detail

import (
	"github.com/poteto0/go-nba-sdk/types"
)

func GetKawaiiPrefix(stats types.PlayerBoxScoreStatistic) string {
	if isTripleDouble(stats) {
		return "👑"
	}

	if isFiveByFive(stats) {
		return "💯"
	}

	if isSniper(stats) {
		return "🎯"
	}

	if isNinja(stats) {
		return "🥷"
	}

	if isBlocker(stats) {
		return "🔒"
	}

	if isAlien(stats) {
		return "👽"
	}

	if isAssister(stats) {
		return "🤝"
	}

	if isMuscle(stats) {
		return "💪"
	}

	return ""
}
