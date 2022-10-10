package main

import (
	"fmt"
	"log"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	table "github.com/evertras/bubble-table/table"
	"github.com/google/go-github/github"
)

type Model struct {
	table    table.Model
	width    int
	margin   int
	Quitting bool
	Data     []*github.PullRequest
	Selected string
}

func AddColumn(key string, header string, size int) table.Column {
	return table.NewFlexColumn(key, header, size).WithStyle(lipgloss.NewStyle().Align(lipgloss.Left))
}

func NewModel(rows []table.Row) Model {
	columns := []table.Column{
		AddColumn("time", "Time", 1),
		AddColumn("branch", "Branch", 3),
		AddColumn("title", "Title", 5),
	}

	keys := table.DefaultKeyMap()
	model := Model{
		table: table.New(columns).WithKeyMap(keys).WithRows(rows).SelectableRows(true).Focused(true),
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
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	m.table, cmd = m.table.Update(msg)
	cmds = append(cmds, cmd)

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
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.recalculateTable()
	}

	cmds = append(cmds, tea.EnterAltScreen)

	return m, tea.Batch(cmds...)
}

func (m Model) View() string {
	body := strings.Builder{}

	if m.Selected != "" {
		// return "\n   HEWWOOO\n\n"V
		comments, err := GetReviewComments(m.Selected)
		// fmt.Println(len(comments))
		if err != nil {
			return fmt.Sprintf("%v\n", err)
		}

		if len(comments) == 0 {
			return fmt.Sprintf("NOPE NOTHING ON %s\n", m.Selected)
		}
		commentsBody := ""
		for _, comment := range comments {
			commentsBody += fmt.Sprintf("%s - %s\n%s\n%s\n", comment.UpdatedAt, comment.Path, comment.From.Username, comment.Body)
		}
		return commentsBody
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
