package game_detail

import (
	"fmt"
	"nba-tui/internal/ui/styles"
	"nba-tui/internal/utils"
	"os/exec"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
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

type Config struct {
	NoDecoration bool
	KawaiiMode   bool
}

type Model struct {
	client            NbaClient
	gameID            string
	boxScore          types.LiveBoxScoreResponse
	pbp               types.LivePlayByPlayResponse
	showingHome       bool
	lastUpdated       time.Time
	focus             focusArea
	logOffset         int
	boxOffset         int
	boxScrollX        int
	selectedPeriod    int
	width             int
	height            int
	OpenBrowser       func(string) error
	config            Config
	searchInput       textinput.Model
	searchMode        bool
	matchedIndices    []int
	currentMatchIndex int
}

func New(client NbaClient, gameID string, config Config) Model {
	ti := textinput.New()
	ti.Placeholder = "Search..."
	ti.Prompt = "/"
	ti.CharLimit = 156
	ti.Width = 30

	return Model{
		client:         client,
		gameID:         gameID,
		showingHome:    true,
		selectedPeriod: 1,
		OpenBrowser: func(url string) error {
			return exec.Command("xdg-open", url).Start()
		},
		config:      config,
		searchInput: ti,
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

func (m Model) getVisibleActions() []types.Action {
	team := m.getCurrentTeam()
	var filteredActions []types.Action
	for _, action := range m.pbp.Game.Actions {
		if action.Period == m.selectedPeriod && action.TeamID == team.TeamId {
			filteredActions = append(filteredActions, action)
		}
	}
	return filteredActions
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.searchMode {
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.Type {
			case tea.KeyEnter:
				m.searchMode = false
				query := m.searchInput.Value()
				actions := m.getVisibleActions()
				m.matchedIndices = SearchActions(actions, query)
				if len(m.matchedIndices) > 0 {
					m.currentMatchIndex = 0
					m.logOffset = m.matchedIndices[0]
					// Auto-switch to game log focus if search found something
					m.focus = gameLogFocus
				}
				return m, nil
			case tea.KeyEsc:
				m.searchMode = false
				m.searchInput.Blur()
				return m, nil
			}
		default:
			var cmd tea.Cmd
			m.searchInput, cmd = m.searchInput.Update(msg)
			return m, cmd
		}
	}

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
		case "/":
			m.searchMode = true
			m.searchInput.Focus()
			m.searchInput.SetValue("")
			return m, nil
		case "n":
			if len(m.matchedIndices) > 0 {
				m.currentMatchIndex++
				if m.currentMatchIndex >= len(m.matchedIndices) {
					m.currentMatchIndex = 0
				}
				m.logOffset = m.matchedIndices[m.currentMatchIndex]
				m.focus = gameLogFocus
			}
		case "N":
			if len(m.matchedIndices) > 0 {
				m.currentMatchIndex--
				if m.currentMatchIndex < 0 {
					m.currentMatchIndex = len(m.matchedIndices) - 1
				}
				m.logOffset = m.matchedIndices[m.currentMatchIndex]
				m.focus = gameLogFocus
			}
		case "ctrl+s":
			m.showingHome = !m.showingHome
			m.logOffset = 0
			m.matchedIndices = []int{}
			m.currentMatchIndex = 0
		case "ctrl+c":
			return m, tea.Quit
		case "ctrl+q":
			m.selectedPeriod++
			if m.selectedPeriod > 4 {
				m.selectedPeriod = 1
			}
			m.logOffset = 0
			m.matchedIndices = []int{}
			m.currentMatchIndex = 0
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
				filteredActions := m.getVisibleActions()
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
	teamInfo := fmt.Sprintf("Selected Team: %s", team.TeamTricode)
	if !m.lastUpdated.IsZero() {
		updateTimeStr := m.lastUpdated.Format("15:04:05") // HH:MM:SS
		teamInfo = fmt.Sprintf("%s (Last Updated: %s)", teamInfo, updateTimeStr)
	}
	selectedTeamView := styles.UnderlineStyle.Render(teamInfo)

	// Render footer first to know its height
	var footerView string
	if m.searchMode {
		footerView = m.searchInput.View()
	} else {
		footerView = m.renderFooter(m.width)
	}

	h_footer := lipgloss.Height(footerView)
	footerView = lipgloss.NewStyle().Width(m.width).Height(h_footer).MaxHeight(h_footer).Render(footerView)

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

	switch {
	case !game.IsGameStart():
		status = "not started"
	case game.IsFinished():
		status = game.GameStatusText
	default:
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

	filteredActions := m.getVisibleActions()

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

			line := fmt.Sprintf("% -5s|%s", action.Clock, desc)

			// Highlight matching rows
			for _, matchIdx := range m.matchedIndices {
				if matchIdx == idx {
					// Check if it's the currently selected match
					if idx == m.matchedIndices[m.currentMatchIndex] {
						// Maybe distinct highlight for current match?
						line = styles.HighlightStyle.Bold(true).Render(line)
					} else {
						line = styles.HighlightStyle.Render(line)
					}
					break
				}
			}

			logLines = append(logLines, line)
		}
	}
	gameLogBody := strings.Join(logLines, "\n")
	return gameLogHeader + "\n" + periodSelector + "\n" + gameLogBody
}

func (m Model) renderBoxScore(team types.Team, width, height int) string {
	s := ""

	// Define columns configuration
	type col struct {
		name  string
		width int
		align lipgloss.Position
	}
	cols := []col{
		{"PLAYER", 15, lipgloss.Left},
		{"MIN", 5, lipgloss.Left},
		{"FGM", 3, lipgloss.Right},
		{"FGA", 3, lipgloss.Right},
		{"FG%", 5, lipgloss.Right},
		{"3PM", 3, lipgloss.Right},
		{"3PA", 3, lipgloss.Right},
		{"3P%", 5, lipgloss.Right},
		{"FTM", 3, lipgloss.Right},
		{"FTA", 3, lipgloss.Right},
		{"FT%", 5, lipgloss.Right},
		{"OREB", 4, lipgloss.Right},
		{"DREB", 4, lipgloss.Right},
		{"REB", 3, lipgloss.Right},
		{"AST", 3, lipgloss.Right},
		{"STL", 3, lipgloss.Right},
		{"BLK", 3, lipgloss.Right},
		{"TO", 3, lipgloss.Right},
		{"PF", 3, lipgloss.Right},
		{"PTS", 3, lipgloss.Right},
		{"+/-", 4, lipgloss.Right},
	}

	renderRow := func(vals []string) string {
		parts := make([]string, 0, len(cols))
		for i, c := range cols {
			val := ""
			if i < len(vals) {
				val = vals[i]
			}
			// Use lipgloss to align and pad. It handles ANSI width correctly.
			cell := lipgloss.NewStyle().Width(c.width).Align(c.align).Render(val)
			parts = append(parts, cell)
		}
		return strings.Join(parts, " ")
	}

	// Header
	headerVals := make([]string, len(cols))
	for i, c := range cols {
		headerVals[i] = c.name
	}
	fullHeader := renderRow(headerVals)

	// Apply horizontal scroll to header
	header := styles.TableHeaderStyle.Render(m.scrollLine(fullHeader, width))
	s += header + "\n"

	if team.Players == nil {
		return s + "No player data"
	}
	players := *team.Players

	// Calculate team highs
	maxPts, maxReb, maxAst := -1, -1, -1
	for _, p := range players {
		if p.Statistics != nil {
			stats := *p.Statistics
			if stats.Pts != nil && *stats.Pts > maxPts {
				maxPts = *stats.Pts
			}
			if stats.Reb != nil && *stats.Reb > maxReb {
				maxReb = *stats.Reb
			}
			if stats.Ast != nil && *stats.Ast > maxAst {
				maxAst = *stats.Ast
			}
		}
	}

	// Reserve space for Total row (Separator + Stats) if available
	hasTeamStats := team.Statistics != nil
	reservedBottom := 0
	if hasTeamStats {
		reservedBottom = 2
	}

	// Height calculation: Header takes 2 lines (due to border?).
	// Actually styles.TableHeaderStyle adds a bottom border, so it consumes vertical space?
	// Render(str) -> content + border.
	// If str is 1 line, result is 2 lines (content + border).
	// So Header consumes 2 lines.
	// We removed Title, so we have more space.
	bodyHeight := height - 2 - reservedBottom
	if bodyHeight < 0 {
		bodyHeight = 0
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

			if p.Statistics == nil {
				// Empty row
				vals := make([]string, len(cols))
				vals[0] = name
				vals[1] = "-"
				s += m.scrollLine(renderRow(vals), width) + "\n"
				continue
			}
			stats := *p.Statistics

			// Kawaii Mode Prefix
			if m.config.KawaiiMode {
				prefix := GetKawaiiPrefix(stats)
				if prefix != "" {
					name = prefix + name
				}
			}

			// Truncate name if too long for column (manual check for safety, though lipgloss handles it)
			if len(name) > 15 {
				name = name[:15]
			}

			clockRaw := stats.MinutesClock()
			min := "-"
			if len(clockRaw) > 5 {
				min = clockRaw[:5]
			}

			// Prepare individual stat strings
			ptsVal := stats.Pts
			rebVal := stats.Reb
			astVal := stats.Ast
			stlVal := stats.Stl
			blkVal := stats.Blk
			pmVal := stats.PlusMinus

			// Initialize base styles for each stat
			ptsStyle := lipgloss.NewStyle()
			rebStyle := lipgloss.NewStyle()
			astStyle := lipgloss.NewStyle()
			stlStyle := lipgloss.NewStyle()
			blkStyle := lipgloss.NewStyle()
			pmStyle := lipgloss.NewStyle()

			if !m.config.NoDecoration { // Apply Team High bolding and Plus/Minus colors
				if ptsVal != nil && *ptsVal == maxPts && maxPts > 0 {
					ptsStyle = ptsStyle.Bold(true)
				}
				if rebVal != nil && *rebVal == maxReb && maxReb > 0 {
					rebStyle = rebStyle.Bold(true)
				}
				if astVal != nil && *astVal == maxAst && maxAst > 0 {
					astStyle = astStyle.Bold(true)
				}

				if pmVal != nil {
					if *pmVal > 0 {
						pmStyle = pmStyle.Foreground(lipgloss.Color("2")) // Green
					} else if *pmVal < 0 {
						pmStyle = pmStyle.Foreground(lipgloss.Color("1")) // Red
					}
				}
			}

			if m.config.KawaiiMode { // Apply Underlining
				if ShouldUnderlineStat("PTS", ptsVal) {
					ptsStyle = ptsStyle.Underline(true)
				}
				if ShouldUnderlineStat("REB", rebVal) {
					rebStyle = rebStyle.Underline(true)
				}
				if ShouldUnderlineStat("AST", astVal) {
					astStyle = astStyle.Underline(true)
				}
				if ShouldUnderlineStat("STL", stlVal) {
					stlStyle = stlStyle.Underline(true)
				}
				if ShouldUnderlineStat("BLK", blkVal) {
					blkStyle = blkStyle.Underline(true)
				}
			}

			// Render all stats with combined styles
			ptsStr := ptsStyle.Render(utils.PtrToIntStr(ptsVal))
			rebStr := rebStyle.Render(utils.PtrToIntStr(rebVal))
			astStr := astStyle.Render(utils.PtrToIntStr(astVal))
			stlStr := stlStyle.Render(utils.PtrToIntStr(stlVal))
			blkStr := blkStyle.Render(utils.PtrToIntStr(blkVal))
			pmStr := pmStyle.Render(utils.PtrToFloatStr2f(pmVal))

			// Construct values matching cols order
			rowVals := []string{
				name,
				min,
				utils.PtrToIntStr(stats.FgM),
				utils.PtrToIntStr(stats.FgA),
				utils.PtrToPctStr(stats.FgPct),
				utils.PtrToIntStr(stats.Fg3M),
				utils.PtrToIntStr(stats.Fg3A),
				utils.PtrToPctStr(stats.Fg3Pct),
				utils.PtrToIntStr(stats.FtM),
				utils.PtrToIntStr(stats.FtA),
				utils.PtrToPctStr(stats.FtPct),
				utils.PtrToIntStr(stats.OReb),
				utils.PtrToIntStr(stats.DReb),
				rebStr,
				astStr,
				stlStr,
				blkStr,
				utils.PtrToIntStr(stats.Tov),
				utils.PtrToIntStr(stats.PF),
				ptsStr,
				pmStr,
			}

			s += m.scrollLine(renderRow(rowVals), width) + "\n"
		}
	}

	if hasTeamStats {
		linesRendered := bodyHeight
		if len(players)-m.boxOffset < bodyHeight {
			linesRendered = len(players) - m.boxOffset
		}
		if linesRendered < 0 {
			linesRendered = 0
		}

		padding := bodyHeight - linesRendered
		if padding > 0 {
			s += strings.Repeat("\n", padding)
		}

		// Separator
		separator := strings.Repeat("─", width)
		s += separator + "\n"

		stats := *team.Statistics

		min := "-"
		if stats.Minutes != "" {
			min = stats.MinutesClock()
			if len(min) > 5 {
				min = min[:5]
			}
		}

		totalVals := []string{
			"TOTAL", min,
			utils.PtrToIntStr(stats.FgM),
			utils.PtrToIntStr(stats.FgA),
			utils.PtrToPctStr(stats.FgPct),
			utils.PtrToIntStr(stats.Fg3M),
			utils.PtrToIntStr(stats.Fg3A),
			utils.PtrToPctStr(stats.Fg3Pct),
			utils.PtrToIntStr(stats.FtM),
			utils.PtrToIntStr(stats.FtA),
			utils.PtrToPctStr(stats.FtPct),
			utils.PtrToIntStr(stats.OReb),
			utils.PtrToIntStr(stats.DReb),
			utils.PtrToIntStr(stats.Reb),
			utils.PtrToIntStr(stats.Ast),
			utils.PtrToIntStr(stats.Stl),
			utils.PtrToIntStr(stats.Blk),
			utils.PtrToIntStr(stats.Tov),
			utils.PtrToIntStr(stats.PF),
			utils.PtrToIntStr(stats.Pts),
			"-",
		}
		s += m.scrollLine(renderRow(totalVals), width)
	}

	return s
}

func (m Model) scrollLine(line string, width int) string {
	const fixedWidth = 16

	// Visual cut for the fixed part (0 to fixedWidth)
	fixed := ansi.Cut(line, 0, fixedWidth)

	remainingWidth := width - fixedWidth
	if remainingWidth <= 0 {
		return fixed
	}

	// Visual cut for the scrollable part
	// Start index is relative to the whole line: fixedWidth + scroll offset
	startVisual := fixedWidth + m.boxScrollX
	endVisual := startVisual + remainingWidth

	scrollable := ansi.Cut(line, startVisual, endVisual)

	return fixed + scrollable
}
