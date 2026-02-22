package game_detail

import (
	"strings"

	"github.com/poteto0/go-nba-sdk/types"
)

func GetKawaiiPrefix(stats types.PlayerBoxScoreStatistic) string {
	var prefixes []string

	// TD(PTS/REB/AST/STL/BLK)の中で3つ以上2桁いってたら`👑`
	doubleDigits := 0
	checkDouble := func(val *int) {
		if val != nil && *val >= 10 {
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
		if val == nil || *val < 5 {
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
	if stats.Fg3M != nil && *stats.Fg3M >= 8 && stats.Fg3Pct != nil && *stats.Fg3Pct >= 0.5 {
		prefixes = append(prefixes, "🎯")
	}

	// `STL>=5`: 🥷🏻
	if stats.Stl != nil && *stats.Stl >= 5 {
		prefixes = append(prefixes, "🥷🏻")
	}

	// `BLK>=7`: 🛡️
	if stats.Blk != nil && *stats.Blk >= 7 {
		prefixes = append(prefixes, "🛡️")
	}

	// `PTS>=50`: 👽
	if stats.Pts != nil && *stats.Pts >= 50 {
		prefixes = append(prefixes, "👽")
	}

	// `AST>=20`: 🤝
	if stats.Ast != nil && *stats.Ast >= 20 {
		prefixes = append(prefixes, "🤝")
	}

	// `REB>=20`: 💪
	if stats.Reb != nil && *stats.Reb >= 20 {
		prefixes = append(prefixes, "💪")
	}

	// Max 3
	if len(prefixes) > 3 {
		prefixes = prefixes[:3]
	}

	return strings.Join(prefixes, "")
}


func ShouldUnderlineStat(statName string, val *int) bool {
	if val == nil {
		return false
	}
	switch statName {
	case "PTS", "REB", "AST":
		return *val >= 10
	case "STL", "BLK":
		return *val > 3
	}
	return false
}
