package game_detail

import (
	"strings"

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

func GetKawaiiPrefix(stats types.PlayerBoxScoreStatistic) string {
	var prefixes []string

	// TD(PTS/REB/AST/STL/BLK)の中で3つ以上2桁いってたら`👑`
	doubleDigits := 0
	checkDouble := func(val *int) {
		if val != nil && *val >= ThresholdDoubleDigits {
			doubleDigits++
		}
	}
	checkDouble(stats.Pts)
	checkDouble(stats.Reb)
	checkDouble(stats.Ast)
	checkDouble(stats.Stl)
	checkDouble(stats.Blk)
	if doubleDigits >= 3 {
		prefixes = append(prefixes, "👑")
	}

	// 5b5(PTS/REB/AST/STL/BLK)が全部5を超えていれば`💯`
	fiveByFive := true
	checkFive := func(val *int) {
		if val == nil || *val < ThresholdFiveByFive {
			fiveByFive = false
		}
	}
	checkFive(stats.Pts)
	checkFive(stats.Reb)
	checkFive(stats.Ast)
	checkFive(stats.Stl)
	checkFive(stats.Blk)
	if fiveByFive {
		prefixes = append(prefixes, "💯")
	}

	// 3PMが8かつ3P%が50%を超えたら`🎯`
	if stats.Fg3M != nil && *stats.Fg3M >= ThresholdSniper3PM && stats.Fg3Pct != nil && *stats.Fg3Pct >= ThresholdSniper3Pct {
		prefixes = append(prefixes, "🎯")
	}

	// `STL>=5`: 🥷🏻
	if stats.Stl != nil && *stats.Stl >= ThresholdSteal {
		prefixes = append(prefixes, "🥷🏻")
	}

	// `BLK>=7`: 🛡️
	if stats.Blk != nil && *stats.Blk >= ThresholdBlock {
		prefixes = append(prefixes, "🛡️")
	}

	// `PTS>=50`: 👽
	if stats.Pts != nil && *stats.Pts >= ThresholdPoints {
		prefixes = append(prefixes, "👽")
	}

	// `AST>=20`: 🤝
	if stats.Ast != nil && *stats.Ast >= ThresholdAssist {
		prefixes = append(prefixes, "🤝")
	}

	// `REB>=20`: 💪
	if stats.Reb != nil && *stats.Reb >= ThresholdRebound {
		prefixes = append(prefixes, "💪")
	}

	// Max 1
	if len(prefixes) > 1 {
		prefixes = prefixes[:1]
	}

	return strings.Join(prefixes, "")
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
