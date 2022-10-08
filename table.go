package main

import (
	"log"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	table "github.com/evertras/bubble-table/table"
	"github.com/google/go-github/github"
)

type Model struct {
	table  table.Model
	width  int
	margin int
}

func AddColumn(key string, header string, size int) table.Column {
	return table.NewFlexColumn(key, header, size).WithStyle(lipgloss.NewStyle().Align(lipgloss.Left))
}

func NewModel(rows []table.Row) Model {
	columns := []table.Column{
		AddColumn("reason", "Reason", 1),
		AddColumn("message", "Message", 3),
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
			cmds = append(cmds, tea.Quit)
		case "enter":

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
	body.WriteString(m.table.View())

	return body.String()
}

func RenderTable(prs []*github.PullRequest) {
	rows := make([]table.Row, 0)

	for _, pr := range prs {
		rows = append(rows, table.NewRow(table.RowData{"reason": pr.Head.GetLabel(), "message": pr.GetTitle()}))
	}

	p := tea.NewProgram(NewModel(rows), tea.WithAltScreen())

	if err := p.Start(); err != nil {
		log.Fatal(err)
	}
}
