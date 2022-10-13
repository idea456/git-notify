package main

import (
	"fmt"
	"log"
	"strings"
	"time"

	markdown "github.com/MichaelMure/go-term-markdown"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	table "github.com/evertras/bubble-table/table"
	"github.com/google/go-github/github"
)

var PAGE_SIZE int = 35
var docStyle = lipgloss.NewStyle().Margin(1, 2)

type Model struct {
	table    table.Model
	list     list.Model
	width    int
	margin   int
	Quitting bool
	Data     []*github.PullRequest
	Selected string
	Loading  bool // use this later to show loading view when transitioning between views
}

type ListItem struct {
	title, description string
}

func (li ListItem) Title() string       { return li.title }
func (li ListItem) Description() string { return li.description }
func (li ListItem) FilterValue() string { return li.Title() }

func AddColumn(key string, header string, size int) table.Column {
	return table.NewFlexColumn(key, header, size).WithStyle(lipgloss.NewStyle().Align(lipgloss.Left))
}

func NewModel(rows []table.Row) Model {
	columns := []table.Column{
		AddColumn("time", "Time", 1),
		AddColumn("branch", "Branch", 3),
		AddColumn("title", "Title", 5),
	}

	var baseStyle = lipgloss.NewStyle().
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240"))

	keys := table.DefaultKeyMap()
	model := Model{
		table: table.New(columns).WithBaseStyle(baseStyle).WithKeyMap(keys).WithRows(rows).SelectableRows(true).Focused(true).WithFooterVisibility(true).WithStaticFooter(time.Now().String()).WithPageSize(PAGE_SIZE).SortByAsc("time"),
		list:  list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 50),
	}

	return model
}

func (m Model) recalculateTable() {
	m.table = m.table.WithTargetWidth(m.width - m.margin)
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd     tea.Cmd
		listCmd tea.Cmd
		cmds    []tea.Cmd
	)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc", "q":
			m.Quitting = true
			cmds = append(cmds, tea.Quit)
		case "left":
			m.Selected = ""
		case "enter":
			m.Selected = m.table.HighlightedRow().Data["ReviewCommentsURL"].(string)
			comments, _ := GetReviewComments(m.Selected)
			commentsBody := make([]list.Item, 0)
			for _, comment := range comments {
				commentsBody = append(commentsBody, ListItem{
					title:       fmt.Sprintf("[%s] %s", comment.From.Username, comment.Path),
					description: string(markdown.Render(string(comment.Body), 80, 6)),
				})
			}
			m.list.Title = m.table.HighlightedRow().Data["title"].(string)
			m.list.SetItems(commentsBody)
			cmds = append(cmds, tea.EnterAltScreen)
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.table.WithPageSize(PAGE_SIZE)
		h, v := docStyle.GetFrameSize()
		m.list.SetSize(msg.Width-h, msg.Height-v)
		m.recalculateTable()
	}

	m.table, cmd = m.table.Update(msg)
	m.list, listCmd = m.list.Update(msg)
	cmds = append(cmds, cmd, listCmd)

	// cmds = append(cmds, tea.EnterAltScreen)

	return m, tea.Batch(cmds...)
}

func (m Model) View() string {
	body := strings.Builder{}

	if m.Selected != "" {
		// comments, err := GetReviewComments(m.Selected)
		// if err != nil {
		// 	return fmt.Sprintf("%v\n", err)
		// }

		// if len(comments) == 0 {
		// 	return fmt.Sprintf("NOPE NOTHING ON %s\n", m.Selected)
		// }
		// commentsBody := make([]list.Item, 0)
		// for _, comment := range comments {
		// 	commentsBody = append(commentsBody, ListItem{
		// 		title:       fmt.Sprintf("[%s] %s", comment.UpdatedAt, comment.From.Username),
		// 		description: comment.Path,
		// 	})
		// }

		// m.list.SetItems(commentsBody)

		body.WriteString(docStyle.Render(m.list.View()))
	} else {
		body.WriteString(m.table.View())
	}

	if m.Quitting {
		return "\n  Bai bai!\n\n"
	}

	return body.String()
}

func RenderTable(prs []*github.PullRequest) {
	rows := make([]table.Row, 0)

	for _, pr := range prs {
		t := pr.GetUpdatedAt()
		rows = append(rows, table.NewRow(table.RowData{
			"time":              fmt.Sprintf("%s %s", t.Weekday(), t.Format("03:04 PM")),
			"branch":            pr.GetHead().GetLabel(),
			"title":             pr.GetTitle(),
			"ReviewCommentsURL": *pr.ReviewCommentsURL,
		}))
	}

	m := NewModel(rows)
	m.Data = prs
	p := tea.NewProgram(m, tea.WithAltScreen())

	if err := p.Start(); err != nil {
		log.Fatal(err)
	}
}
