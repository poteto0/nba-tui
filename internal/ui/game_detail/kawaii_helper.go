package game_detail

import (
	"github.com/poteto0/go-nba-sdk/types"
)

const (
	ThresholdDoubleDigits = 10
	ThresholdFiveByFive   = 5
	ThresholdSniper3PM    = 8
	ThresholdSniper3Pct   = 0.5
	ThresholdSteal        = 5
	ThresholdBlock        = 7
	ThresholdPoints       = 50
	ThresholdAssist       = 20
	ThresholdRebound      = 20

	ThresholdUnderlineHigh = 10
	ThresholdUnderlineLow  = 3
)

func isTripleDouble(stats types.PlayerBoxScoreStatistic) bool {
	doubleDigitsCount := 0
	if isDoubleDigits(stats.Pts) {
		doubleDigitsCount++
	}

	if isDoubleDigits(stats.Reb) {
		doubleDigitsCount++
	}

	if isDoubleDigits(stats.Ast) {
		doubleDigitsCount++
	}

	if isDoubleDigits(stats.Stl) {
		doubleDigitsCount++
	}

	if isDoubleDigits(stats.Blk) {
		doubleDigitsCount++
	}
	return doubleDigitsCount >= 3
}

func isDoubleDigits(val *int) bool {
	if val == nil {
		return false
	}

	return *val >= ThresholdDoubleDigits
}

func isFiveByFive(stats types.PlayerBoxScoreStatistic) bool {
	fiveByFiveCount := 0
	if isFive(stats.Pts) {
		fiveByFiveCount++
	}
	if isFive(stats.Reb) {
		fiveByFiveCount++
	}
	if isFive(stats.Ast) {
		fiveByFiveCount++
	}
	if isFive(stats.Stl) {
		fiveByFiveCount++
	}
	if isFive(stats.Blk) {
		fiveByFiveCount++
	}
	return fiveByFiveCount >= 5
}

func isFive(val *int) bool {
	if val == nil {
		return false
	}
	return *val >= ThresholdFiveByFive
}

func isSniper(stats types.PlayerBoxScoreStatistic) bool {
	return stats.Fg3M != nil && *stats.Fg3M >= ThresholdSniper3PM && stats.Fg3Pct != nil && *stats.Fg3Pct >= ThresholdSniper3Pct
}

func isNinja(stats types.PlayerBoxScoreStatistic) bool {
	return stats.Stl != nil && *stats.Stl >= ThresholdSteal
}

func isBlocker(stats types.PlayerBoxScoreStatistic) bool {
	return stats.Blk != nil && *stats.Blk >= ThresholdBlock
}

func isAlien(stats types.PlayerBoxScoreStatistic) bool {
	return stats.Pts != nil && *stats.Pts >= ThresholdPoints
}

func isAssister(stats types.PlayerBoxScoreStatistic) bool {
	return stats.Ast != nil && *stats.Ast >= ThresholdAssist
}

func isMuscle(stats types.PlayerBoxScoreStatistic) bool {
	return stats.Reb != nil && *stats.Reb >= ThresholdRebound
}

func ShouldUnderlineStat(statName string, val *int) bool {
	if val == nil {
		return false
	}
	switch statName {
	case "PTS", "REB", "AST":
		return *val >= ThresholdUnderlineHigh
	case "STL", "BLK":
		return *val > ThresholdUnderlineLow
	}
	return false
}
