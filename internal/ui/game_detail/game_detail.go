package game_detail

import (
	"fmt"
	"nba-tui/internal/ui/styles"
	"os/exec"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/poteto0/go-nba-sdk/types"
)

type focusArea int

const (
	boxScoreFocus focusArea = iota
	gameLogFocus
)

type NbaClient interface {
	GetBoxScore(gameID string) (types.LiveBoxScoreResponse, error)
	GetPlayByPlay(gameID string) (types.LivePlayByPlayResponse, error)
}

type Model struct {
	client         NbaClient
	gameID         string
	boxScore       types.LiveBoxScoreResponse
	pbp            types.LivePlayByPlayResponse
	showingHome    bool
	lastUpdated    time.Time
	focus          focusArea
	logOffset      int
	boxOffset      int
	boxScrollX     int
	selectedPeriod int
	width          int
	height         int
	OpenBrowser    func(string) error
}

func New(client NbaClient, gameID string) Model {
	return Model{
		client:         client,
		gameID:         gameID,
		showingHome:    true,
		selectedPeriod: 1,
		OpenBrowser: func(url string) error {
			return exec.Command("xdg-open", url).Start()
		},
	}
}

func (m *Model) SetLastUpdated(t time.Time) {
	m.lastUpdated = t
}

func (m Model) IsShowingHome() bool {
	return m.showingHome
}

func (m Model) GetFocus() int {
	return int(m.focus)
}

func (m Model) GetLogOffset() int {
	return m.logOffset
}

func (m Model) GetBoxOffset() int {
	return m.boxOffset
}

func (m Model) GetSelectedPeriod() int {
	return m.selectedPeriod
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(
		m.fetchBoxScore,
		m.fetchPlayByPlay,
	)
}

func (m Model) fetchBoxScore() tea.Msg {
	res, err := m.client.GetBoxScore(m.gameID)
	if err != nil {
		return ErrorMsg(err)
	}
	return BoxScoreMsg(res)
}

func (m Model) fetchPlayByPlay() tea.Msg {
	res, err := m.client.GetPlayByPlay(m.gameID)
	if err != nil {
		return ErrorMsg(err)
	}
	return PlayByPlayMsg(res)
}

type BoxScoreMsg types.LiveBoxScoreResponse
type PlayByPlayMsg types.LivePlayByPlayResponse
type ErrorMsg error

func (m Model) getCurrentTeam() types.Team {
	if m.showingHome {
		return m.boxScore.Game.HomeTeam
	}
	return m.boxScore.Game.AwayTeam
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case BoxScoreMsg:
		m.boxScore = types.LiveBoxScoreResponse(msg)
		m.lastUpdated = time.Now()

	case PlayByPlayMsg:
		m.pbp = types.LivePlayByPlayResponse(msg)
		m.lastUpdated = time.Now()

	case tea.KeyMsg:
		team := m.getCurrentTeam()
		switch msg.String() {
		case "ctrl+s":
			m.showingHome = !m.showingHome
			m.logOffset = 0
		case "ctrl+c":
			return m, tea.Quit
		case "ctrl+q":
			m.selectedPeriod++
			if m.selectedPeriod > 4 {
				m.selectedPeriod = 1
			}
			m.logOffset = 0
		case "ctrl+w":
			url := fmt.Sprintf("https://www.nba.com/game/%s", m.gameID)
			if m.OpenBrowser != nil {
				_ = m.OpenBrowser(url)
			}
		case "ctrl+b":
			m.focus = boxScoreFocus
		case "ctrl+l":
			m.focus = gameLogFocus
		case "h", "left":
			if m.focus == boxScoreFocus {
				if m.boxScrollX > 0 {
					m.boxScrollX--
				}
			}
		case "l", "right":
			if m.focus == boxScoreFocus {
				headerFormat := "%-15s %-5s %3s %3s %5s %3s %3s %5s %3s %3s %5s %4s %4s %3s %3s %3s %3s %3s %3s %3s %4s"
				fullHeader := fmt.Sprintf(headerFormat,
					"PLAYER", "MIN", "FGM", "FGA", "FG%", "3PM", "3PA", "3P%", "FTM", "FTA", "FT%", "OREB", "DREB", "REB", "AST", "STL", "BLK", "TO", "PF", "PTS", "+/-")
				
				w_boxscore := (m.width * 6) / 10
				if m.width < 100 {
					w_boxscore = m.width
				}
				contentWidth := w_boxscore - 2
				
				maxScroll := len(fullHeader) - contentWidth
				if maxScroll < 0 {
					maxScroll = 0
				}
				
				if m.boxScrollX < maxScroll {
					m.boxScrollX++
				}
			}
		case "j", "down":
			if m.focus == boxScoreFocus {
				if team.Players != nil {
					players := *team.Players
					if m.boxOffset < len(players)-1 {
						m.boxOffset++
					}
				}
			} else {
				var filteredActions []types.Action
				for _, action := range m.pbp.Game.Actions {
					if action.Period == m.selectedPeriod && action.TeamID == team.TeamId {
						filteredActions = append(filteredActions, action)
					}
				}
				if m.logOffset < len(filteredActions)-1 {
					m.logOffset++
				}
			}
		case "k", "up":
			if m.focus == boxScoreFocus {
				if m.boxOffset > 0 {
					m.boxOffset--
				}
			} else {
				if m.logOffset > 0 {
					m.logOffset--
				}
			}
		}
	}
	return m, nil
}

func (m Model) View() string {
	if m.boxScore.Game.GameId == "" {
		return "Loading..."
	}

	if m.width < 30 || m.height < 10 {
		return "Terminal too small. Please enlarge."
	}

	team := m.getCurrentTeam()

	// 1. Fixed heights
	h_selected := 1
	selectedTeamView := styles.UnderlineStyle.Render(fmt.Sprintf("Selected Team: %s", team.TeamTricode))

	// Render footer first to know its height
	footerRaw := m.renderFooter(m.width)
	h_footer := lipgloss.Height(footerRaw)
	footerView := lipgloss.NewStyle().Width(m.width).Height(h_footer).MaxHeight(h_footer).Render(footerRaw)

	// 2. Allocate remaining height based on ratios
	// Available for Header + Main
	h_available := m.height - h_selected - h_footer

	h_unit := h_available / 9 // 1 (header) + 8 (main)
	if h_unit < 1 {
		h_unit = 1
	}

	h_header_box := h_unit
	if h_header_box < 4 {
		h_header_box = 4
	}
	// Cap header height if terminal is very small
	if h_header_box > h_available-2 {
		h_header_box = h_available - 2
		if h_header_box < 2 {
			h_header_box = 2
		}
	}

	h_main := h_available - h_header_box
	if h_main < 0 {
		h_main = 0
	}

	headerStr := m.renderHeaderStr()
	var headerBox string
	var mainView string

	if m.width >= 100 {
		// Horizontal Layout: widths 6:4
		w_boxscore := (m.width * 6) / 10
		w_gamelog := m.width - w_boxscore

		if h_main >= 4 {
			bsContent := m.renderBoxScore(team, w_boxscore-2, h_main-2)
			bsStyle := styles.BorderStyle
			if m.focus == boxScoreFocus {
				bsStyle = styles.ActiveBorderStyle
			}
			boxScore := bsStyle.Width(w_boxscore).Height(h_main).MaxHeight(h_main).Render(bsContent)

			glContent := m.renderGameLog(w_gamelog-2, h_main-2)
			glStyle := styles.BorderStyle
			if m.focus == gameLogFocus {
				glStyle = styles.ActiveBorderStyle
			}
			gameLog := glStyle.Width(w_gamelog).Height(h_main).MaxHeight(h_main).Render(glContent)

			mainView = lipgloss.JoinHorizontal(lipgloss.Top, boxScore, gameLog)
		}

		headerBox = styles.BorderStyle.Width(m.width).Height(h_header_box).MaxHeight(h_header_box).Align(lipgloss.Center, lipgloss.Center).Render(headerStr)
	} else {
		// Vertical Layout: heights 4:4 split of h_main
		if h_main >= 6 {
			h_boxscore := h_main / 2
			h_gamelog := h_main - h_boxscore

			bsContent := m.renderBoxScore(team, m.width-2, h_boxscore-2)
			bsStyle := styles.BorderStyle
			if m.focus == boxScoreFocus {
				bsStyle = styles.ActiveBorderStyle
			}
			boxScore := bsStyle.Width(m.width).Height(h_boxscore).MaxHeight(h_boxscore).Render(bsContent)

			glContent := m.renderGameLog(m.width-2, h_gamelog-2)
			glStyle := styles.BorderStyle
			if m.focus == gameLogFocus {
				glStyle = styles.ActiveBorderStyle
			}
			gameLog := glStyle.Width(m.width).Height(h_gamelog).MaxHeight(h_gamelog).Render(glContent)

			mainView = lipgloss.JoinVertical(lipgloss.Left, boxScore, gameLog)
		}

		headerBox = styles.BorderStyle.Width(m.width).Height(h_header_box).MaxHeight(h_header_box).Align(lipgloss.Center, lipgloss.Center).Render(headerStr)
	}

	if mainView == "" {
		return lipgloss.JoinVertical(lipgloss.Left, selectedTeamView, headerBox, footerView)
	}
	return lipgloss.JoinVertical(lipgloss.Left, selectedTeamView, headerBox, mainView, footerView)
}

func (m Model) renderHeaderStr() string {
	game := m.boxScore.Game
	var status string
	
	if !game.IsGameStart() {
		status = "not started"
	} else if game.IsFinished() {
		status = game.GameStatusText
	} else {
		// In-progress
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
		
		status = fmt.Sprintf("%s (%s)", periodStr, clock)
	}

	homeTricode := game.HomeTeam.TeamTricode
	homeScore := fmt.Sprintf("%d", game.HomeTeam.Score)
	awayTricode := game.AwayTeam.TeamTricode
	awayScore := fmt.Sprintf("%d", game.AwayTeam.Score)

	if game.HomeTeam.Score > game.AwayTeam.Score {
		homeTricode = styles.BoldStyle.Render(homeTricode)
		homeScore = styles.BoldStyle.Render(homeScore)
	} else if game.AwayTeam.Score > game.HomeTeam.Score {
		awayTricode = styles.BoldStyle.Render(awayTricode)
		awayScore = styles.BoldStyle.Render(awayScore)
	}

	return fmt.Sprintf("%s\n%s (%s) | %s (%s)",
		status,
		homeTricode, homeScore,
		awayTricode, awayScore,
	)
}

func (m Model) renderFooter(width int) string {
	helpText := "<hjkli←↓↑→ >: move, <ctrl+s>: switch team, <ctrl+b>: box, <ctrl+l>: log, <ctrl+q>: period, <ctrl+w>: watch, <ctrl+c>: quit"
	var footerText string
	if !m.lastUpdated.IsZero() {
		footerText = fmt.Sprintf("Last updated: %s\n%s", m.lastUpdated.Format(time.RFC1123), helpText)
	} else {
		footerText = helpText
	}
	// Truncate footer if it's too wide to prevent wrapping
	if len(footerText) > width {
		// Very basic truncation for safety
		lines := strings.Split(footerText, "\n")
		for i, line := range lines {
			if len(line) > width {
				lines[i] = line[:width-3] + "..."
			}
		}
		footerText = strings.Join(lines, "\n")
	}
	return footerText
}

func (m Model) renderGameLog(width, height int) string {
	if height < 3 {
		return ""
	}
	// Period Selector
	periods := []string{"1Q", "2Q", "3Q", "4Q"}
	var selectorParts []string
	for i, p := range periods {
		pNum := i + 1
		if pNum == m.selectedPeriod {
			selectorParts = append(selectorParts, styles.UnderlineStyle.Render(p))
		} else {
			selectorParts = append(selectorParts, styles.FaintStyle.Render(p))
		}
	}
	periodSelectorContent := strings.Join(selectorParts, " | ")
	periodSelector := lipgloss.NewStyle().Width(width).Align(lipgloss.Center).Render(periodSelectorContent)

	gameLogHeader := lipgloss.NewStyle().Width(width).Align(lipgloss.Center).Render("gamelog")

	team := m.getCurrentTeam()
	var filteredActions []types.Action
	for _, action := range m.pbp.Game.Actions {
		if action.Period == m.selectedPeriod && action.TeamID == team.TeamId {
			filteredActions = append(filteredActions, action)
		}
	}

	bodyHeight := height - 2 // Minus header and selector
	if bodyHeight < 1 {
		return gameLogHeader + "\n" + periodSelector
	}

	var logLines []string
	for i := 0; i < bodyHeight; i++ {
		idx := m.logOffset + i
		if idx < len(filteredActions) {
			action := filteredActions[idx]
			desc := action.Description
			clockWidth := 6
			descMaxWidth := width - clockWidth
			if len(desc) > descMaxWidth && descMaxWidth > 3 {
				desc = desc[:descMaxWidth-3] + "..."
			}
			logLines = append(logLines, fmt.Sprintf("% -5s|%s", action.Clock, desc))
		}
	}
	gameLogBody := strings.Join(logLines, "\n")
	return gameLogHeader + "\n" + periodSelector + "\n" + gameLogBody
}

func (m Model) renderBoxScore(team types.Team, width, height int) string {
	if height < 4 {
		return "Box Scores"
	}
	s := "Box Scores\n"

	// Format:
	// PLAYER(15) | MIN(5) | Others right aligned
	// %-15s (Left)
	// %-5s  (Left, MIN)
	// %3s   (Right, e.g. PTS, REB)
	// All other stats are numeric or pct, so Right Align (%Xs or %Xd) is standard.
	// Header also needs to match.
	// FGM(3) FGA(3) FG%(5) 3PM(3) 3PA(3) 3P%(5) FTM(3) FTA(3) FT%(5) OREB(4) DREB(4) REB(3) AST(3) STL(3) BLK(3) TO(3) PF(3) PTS(3) +/-(4)

	// Construct format strings
	// Header:
	// PLAYER          MIN   FGM FGA FG%   3PM 3PA 3P%   FTM FTA FT%   OREB DREB REB AST STL BLK TO  PF  PTS +/-
	// Note: spacing between columns.
	// Let's use specific widths.
	// We construct the FULL line first, then slice it.

	headerFormat := "%-15s %-5s %3s %3s %5s %3s %3s %5s %3s %3s %5s %4s %4s %3s %3s %3s %3s %3s %3s %3s %4s"
	fullHeader := fmt.Sprintf(headerFormat,
		"PLAYER", "MIN", "FGM", "FGA", "FG%", "3PM", "3PA", "3P%", "FTM", "FTA", "FT%", "OREB", "DREB", "REB", "AST", "STL", "BLK", "TO", "PF", "PTS", "+/-")

	// Apply horizontal scroll to header
	header := styles.TableHeaderStyle.Render(m.scrollLine(fullHeader, width, 16))
	s += header + "\n"

	if team.Players == nil {
		return s + "No player data"
	}
	players := *team.Players

	bodyHeight := height - 3 // Title + Header(2 lines)
	if bodyHeight < 1 {
		return s
	}

	for i := 0; i < bodyHeight; i++ {
		idx := m.boxOffset + i
		if idx < len(players) {
			p := players[idx]
			name := ""
			if len(p.FirstName) > 0 {
				name = fmt.Sprintf("%s.%s", string(p.FirstName[0]), p.FamilyName)
			} else {
				name = p.FamilyName
			}
			// Truncate name if too long for column
			if len(name) > 15 {
				name = name[:15]
			}

			if p.Statistics == nil {
				// We still need to format the line to respect scrolling, even if empty stats
				// Or just show name and -
				line := fmt.Sprintf("%-15s -", name)
				s += m.scrollLine(line, width, 16) + "\n"
				continue
			}
			stats := *p.Statistics

			getInt := func(i *int) int {
				if i == nil {
					return 0
				}
				return *i
			}
			getPct := func(f *float64) string {
				if f == nil {
					return "0.0"
				}
				return fmt.Sprintf("%.1f", *f*100)
			}
			getFloat := func(f *float64) string {
				if f == nil {
					return "0"
				}
				return fmt.Sprintf("%.0f", *f)
			}

			clockRaw := stats.MinutesClock()
			min := "-"
			if len(clockRaw) > 5 {
				min = clockRaw[:5]
			}

			// Use same widths as header
			// Right align means %3d etc.
			// %-15s (Left)
			// %-5s  (Left)
			// %3d   (Right) ...
			
			// FG% is %5s (Right) to match header width 5
			
			fullLine := fmt.Sprintf(headerFormat,
				name, min,
				fmt.Sprintf("%d", getInt(stats.FgM)),
				fmt.Sprintf("%d", getInt(stats.FgA)),
				getPct(stats.FgPct),
				fmt.Sprintf("%d", getInt(stats.Fg3M)),
				fmt.Sprintf("%d", getInt(stats.Fg3A)),
				getPct(stats.Fg3Pct),
				fmt.Sprintf("%d", getInt(stats.FtM)),
				fmt.Sprintf("%d", getInt(stats.FtA)),
				getPct(stats.FtPct),
				fmt.Sprintf("%d", getInt(stats.OReb)),
				fmt.Sprintf("%d", getInt(stats.DReb)),
				fmt.Sprintf("%d", getInt(stats.Reb)),
				fmt.Sprintf("%d", getInt(stats.Ast)),
				fmt.Sprintf("%d", getInt(stats.Stl)),
				fmt.Sprintf("%d", getInt(stats.Blk)),
				fmt.Sprintf("%d", getInt(stats.Tov)),
				fmt.Sprintf("%d", getInt(stats.PF)),
				fmt.Sprintf("%d", getInt(stats.Pts)),
				getFloat(stats.PlusMinus),
			)

			s += m.scrollLine(fullLine, width, 16) + "\n"
		}
	}
	return s
}

func (m Model) scrollLine(line string, width int, fixedWidth int) string {
	if len(line) <= fixedWidth {
		return line
	}
	
	fixed := line[:fixedWidth]
	scrollable := line[fixedWidth:]
	
	if m.boxScrollX >= len(scrollable) {
		return fixed
	}
	
	remainingWidth := width - fixedWidth
	if remainingWidth <= 0 {
		return fixed[:width]
	}
	
	start := m.boxScrollX
	end := start + remainingWidth
	if end > len(scrollable) {
		end = len(scrollable)
	}
	
	return fixed + scrollable[start:end]
}
