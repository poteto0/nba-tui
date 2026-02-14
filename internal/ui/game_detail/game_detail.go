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
			// Horizontal scroll if needed in future
		case "l", "right":
			// Horizontal scroll if needed in future
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

	// 1. Calculate Vertical Heights based on ratios (Total 10 parts)
	h_unit := m.height / 10
	if h_unit < 1 {
		h_unit = 1
	}

	// Selected Team: Always 1 line
	h_selected := 1
	selectedTeamView := styles.UnderlineStyle.Render(fmt.Sprintf("Selected Team: %s", team.TeamTricode))

	// Header: 1 part (at least 4 lines for content box)
	h_header_box := h_unit
	if h_header_box < 4 {
		h_header_box = 4
	}

	// Footer: 1 part (at least 1 line)
	h_footer := h_unit
	if h_footer < 1 {
		h_footer = 1
	}

	// Main Area (Box Score + Game Log): Remaining height
	h_main := m.height - h_selected - h_header_box - h_footer
	if h_main < 4 {
		// Fallback if terminal is very short
		h_main = 0
		h_header_box = m.height - h_selected - h_footer
		if h_header_box < 3 {
			h_header_box = 3
		}
	}

	headerStr := m.renderHeaderStr()
	var mainView string
	var headerBox string

	if m.width >= 100 {
		// Horizontal Layout: widths 6:4
		w_boxscore := (m.width * 6) / 10
		w_gamelog := m.width - w_boxscore

		if h_main > 0 {
			bsContent := m.renderBoxScore(team, w_boxscore-2, h_main-2)
			bsStyle := styles.BorderStyle
			if m.focus == boxScoreFocus {
				bsStyle = styles.ActiveBorderStyle
			}
			boxScore := bsStyle.Width(w_boxscore).Height(h_main).Render(bsContent)

			glContent := m.renderGameLog(w_gamelog-2, h_main-2)
			glStyle := styles.BorderStyle
			if m.focus == gameLogFocus {
				glStyle = styles.ActiveBorderStyle
			}
			gameLog := glStyle.Width(w_gamelog).Height(h_main).Render(glContent)

			mainView = lipgloss.JoinHorizontal(lipgloss.Top, boxScore, gameLog)
		}

		// Header spans full width
		headerBox = styles.BorderStyle.Width(m.width).Height(h_header_box).Align(lipgloss.Center, lipgloss.Center).Render(headerStr)
	} else {
		// Vertical Layout: heights 4:4 (equal split of h_main)
		if h_main > 0 {
			h_boxscore := h_main / 2
			h_gamelog := h_main - h_boxscore

			bsContent := m.renderBoxScore(team, m.width-2, h_boxscore-2)
			bsStyle := styles.BorderStyle
			if m.focus == boxScoreFocus {
				bsStyle = styles.ActiveBorderStyle
			}
			boxScore := bsStyle.Width(m.width).Height(h_boxscore).Render(bsContent)

			glContent := m.renderGameLog(m.width-2, h_gamelog-2)
			glStyle := styles.BorderStyle
			if m.focus == gameLogFocus {
				glStyle = styles.ActiveBorderStyle
			}
			gameLog := glStyle.Width(m.width).Height(h_gamelog).Render(glContent)

			mainView = lipgloss.JoinVertical(lipgloss.Left, boxScore, gameLog)
		}

		headerBox = styles.BorderStyle.Width(m.width).Height(h_header_box).Align(lipgloss.Center, lipgloss.Center).Render(headerStr)
	}

	footerText := m.renderFooter(m.width)

	if mainView == "" {
		return lipgloss.JoinVertical(lipgloss.Left, selectedTeamView, headerBox, footerText)
	}
	return lipgloss.JoinVertical(lipgloss.Left, selectedTeamView, headerBox, mainView, footerText)
}

func (m Model) renderHeaderStr() string {
	game := m.boxScore.Game
	status := game.GameStatusText
	if !game.IsFinished() {
		status = fmt.Sprintf("%dQ (%s)", game.Period, game.GameClock)
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
	// Note: We don't apply Width here because JoinVertical will align it, 
	// and excessive width wrapping might push content up.
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
	header := styles.TableHeaderStyle.Render(fmt.Sprintf("% -15s % -5s % -3s % -3s % -3s", "PLAYER", "MIN", "PTS", "REB", "AST"))
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
			if p.Statistics == nil {
				s += fmt.Sprintf("% -15s -\n", name)
				continue
			}
			stats := *p.Statistics

			min := stats.Minutes
			pts := 0
			if stats.Pts != nil {
				pts = *stats.Pts
			}
			reb := 0
			if stats.Reb != nil {
				reb = *stats.Reb
			}
			ast := 0
			if stats.Ast != nil {
				ast = *stats.Ast
			}

			line := fmt.Sprintf("% -15s % -5s % -3d % -3d % -3d",
				name, min, pts, reb, ast)
			if len(line) > width {
				line = line[:width]
			}
			s += line + "\n"
		}
	}
	return s
}
