// Package tui implements the step-by-step installation wizard: one step per
// asset category with checkboxes, an extras step for standalone TUI toggles,
// a confirmation step with the install plan, and a final report.
package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"angel-ai-opencode/internal/catalog"
	"angel-ai-opencode/internal/install"
)

var (
	titleStyle    = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("212"))
	stepStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	cursorStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("212"))
	checkedStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("42"))
	helpStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	errorStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
	doneStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("42"))
	selectedCount = lipgloss.NewStyle().Foreground(lipgloss.Color("110"))
)

type phase int

const (
	selecting phase = iota
	extrasPhase
	confirming
	finished
)

type installedMsg struct {
	report []string
	err    error
}

// Model is the Bubble Tea model for the wizard.
type Model struct {
	categories []catalog.Category
	selected   [][]bool

	extras        []install.ExtraOption
	extraSelected []bool

	assetsDir string
	configDir string

	phase  phase
	step   int
	cursor int
	report []string
	err    error
}

// New builds the wizard with every item preselected.
func New(categories []catalog.Category, assetsDir, configDir string) Model {
	selected := make([][]bool, len(categories))
	for i, category := range categories {
		selected[i] = make([]bool, len(category.Items))
		for j := range selected[i] {
			selected[i][j] = true
		}
	}
	extras := install.ExtraOptions
	extraSelected := make([]bool, len(extras))
	for i := range extraSelected {
		extraSelected[i] = true
	}
	return Model{
		categories:    categories,
		selected:      selected,
		extras:        extras,
		extraSelected: extraSelected,
		assetsDir:     assetsDir,
		configDir:     configDir,
	}
}

func (m Model) Init() tea.Cmd { return nil }

func (m Model) chosen() []catalog.Item {
	var items []catalog.Item
	for i, category := range m.categories {
		for j, item := range category.Items {
			if m.selected[i][j] {
				items = append(items, item)
			}
		}
	}
	return items
}

func (m Model) chosenExtras() map[string]bool {
	selected := make(map[string]bool, len(m.extras))
	for i, extra := range m.extras {
		selected[extra.Key] = m.extraSelected[i]
	}
	return selected
}

// totalSteps counts categories + the extras step (if any) + confirmation.
func (m Model) totalSteps() int {
	total := len(m.categories) + 1
	if len(m.extras) > 0 {
		total++
	}
	return total
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case installedMsg:
		m.report = msg.report
		m.err = msg.err
		m.phase = finished
		return m, nil
	case tea.KeyMsg:
		key := msg.String()
		if key == "ctrl+c" || key == "q" {
			return m, tea.Quit
		}
		switch m.phase {
		case selecting:
			return m.updateSelecting(key)
		case extrasPhase:
			return m.updateExtras(key)
		case confirming:
			return m.updateConfirming(key)
		case finished:
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m Model) updateSelecting(key string) (tea.Model, tea.Cmd) {
	items := m.categories[m.step].Items
	switch key {
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
	case "down", "j":
		if m.cursor < len(items)-1 {
			m.cursor++
		}
	case " ":
		m.selected[m.step][m.cursor] = !m.selected[m.step][m.cursor]
	case "a":
		for j := range m.selected[m.step] {
			m.selected[m.step][j] = true
		}
	case "n":
		for j := range m.selected[m.step] {
			m.selected[m.step][j] = false
		}
	case "left", "h", "b":
		if m.step > 0 {
			m.step--
			m.cursor = 0
		}
	case "enter", "right", "l":
		if m.step < len(m.categories)-1 {
			m.step++
			m.cursor = 0
		} else if len(m.extras) > 0 {
			m.phase = extrasPhase
			m.cursor = 0
		} else {
			m.phase = confirming
		}
	}
	return m, nil
}

func (m Model) updateExtras(key string) (tea.Model, tea.Cmd) {
	switch key {
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
	case "down", "j":
		if m.cursor < len(m.extras)-1 {
			m.cursor++
		}
	case " ":
		m.extraSelected[m.cursor] = !m.extraSelected[m.cursor]
	case "a":
		for i := range m.extraSelected {
			m.extraSelected[i] = true
		}
	case "n":
		for i := range m.extraSelected {
			m.extraSelected[i] = false
		}
	case "left", "h", "b":
		m.phase = selecting
		m.step = len(m.categories) - 1
		m.cursor = 0
	case "enter", "right", "l":
		m.phase = confirming
	}
	return m, nil
}

func (m Model) updateConfirming(key string) (tea.Model, tea.Cmd) {
	switch key {
	case "left", "h", "b", "esc":
		if len(m.extras) > 0 {
			m.phase = extrasPhase
			m.cursor = 0
		} else {
			m.phase = selecting
			m.step = len(m.categories) - 1
			m.cursor = 0
		}
	case "enter":
		items := m.chosen()
		extras := m.chosenExtras()
		return m, func() tea.Msg {
			report, err := install.Apply(items, m.configDir)
			extraReport, extraErr := install.ApplyExtras(extras, m.assetsDir, m.configDir)
			report = append(report, extraReport...)
			if err == nil {
				err = extraErr
			}
			return installedMsg{report: report, err: err}
		}
	}
	return m, nil
}

func (m Model) View() string {
	var b strings.Builder
	b.WriteString(titleStyle.Render("Angel AI — instalador de opencode") + "\n\n")

	switch m.phase {
	case selecting:
		category := m.categories[m.step]
		b.WriteString(stepStyle.Render(fmt.Sprintf("Paso %d/%d", m.step+1, m.totalSteps())))
		b.WriteString("  " + titleStyle.Render(category.Title) + "\n\n")
		for j, item := range category.Items {
			cursor := "  "
			if j == m.cursor {
				cursor = cursorStyle.Render("> ")
			}
			check := "[ ]"
			if m.selected[m.step][j] {
				check = checkedStyle.Render("[x]")
			}
			b.WriteString(fmt.Sprintf("%s%s %s\n", cursor, check, item.Name))
		}
		b.WriteString("\n" + helpStyle.Render("espacio marcar · a todos · n ninguno · ←/→ paso · enter siguiente · q salir"))
	case extrasPhase:
		b.WriteString(stepStyle.Render(fmt.Sprintf("Paso %d/%d", len(m.categories)+1, m.totalSteps())))
		b.WriteString("  " + titleStyle.Render("Extras de la TUI") + "\n\n")
		for j, extra := range m.extras {
			cursor := "  "
			if j == m.cursor {
				cursor = cursorStyle.Render("> ")
			}
			check := "[ ]"
			if m.extraSelected[j] {
				check = checkedStyle.Render("[x]")
			}
			b.WriteString(fmt.Sprintf("%s%s %s\n", cursor, check, extra.Label))
			b.WriteString("       " + helpStyle.Render(extra.Description) + "\n")
		}
		b.WriteString("\n" + helpStyle.Render("espacio marcar · a todos · n ninguno · ←/→ paso · enter siguiente · q salir"))
	case confirming:
		b.WriteString(stepStyle.Render(fmt.Sprintf("Paso %d/%d", m.totalSteps(), m.totalSteps())))
		b.WriteString("  " + titleStyle.Render("Confirmar instalación") + "\n\n")
		for i, category := range m.categories {
			count := 0
			for _, on := range m.selected[i] {
				if on {
					count++
				}
			}
			b.WriteString(fmt.Sprintf("  %s: %s\n", category.Title, selectedCount.Render(fmt.Sprintf("%d/%d", count, len(category.Items)))))
		}
		if len(m.extras) > 0 {
			count := 0
			for _, on := range m.extraSelected {
				if on {
					count++
				}
			}
			b.WriteString(fmt.Sprintf("  Extras de la TUI: %s\n", selectedCount.Render(fmt.Sprintf("%d/%d", count, len(m.extras)))))
		}
		plan := install.Plan(m.chosen(), m.configDir)
		plan = append(plan, install.PlanExtras(m.chosenExtras(), m.configDir)...)
		b.WriteString("\n" + stepStyle.Render(fmt.Sprintf("%d acciones sobre %s", len(plan), m.configDir)) + "\n")
		for _, line := range truncate(plan, 12) {
			b.WriteString("  " + line + "\n")
		}
		b.WriteString("\n" + helpStyle.Render("enter instalar · ← volver · q salir"))
	case finished:
		if m.err != nil {
			b.WriteString(errorStyle.Render("Error: "+m.err.Error()) + "\n\n")
		} else {
			b.WriteString(doneStyle.Render("Instalación completada") + "\n\n")
		}
		for _, line := range truncate(m.report, 20) {
			b.WriteString("  " + line + "\n")
		}
		b.WriteString("\n" + helpStyle.Render("pulsa cualquier tecla para salir"))
	}
	b.WriteString("\n")
	return b.String()
}

func truncate(lines []string, max int) []string {
	if len(lines) <= max {
		return lines
	}
	head := make([]string, max, max+1)
	copy(head, lines[:max])
	return append(head, fmt.Sprintf("… y %d más", len(lines)-max))
}

// Run starts the wizard.
func Run(categories []catalog.Category, assetsDir, configDir string) error {
	_, err := tea.NewProgram(New(categories, assetsDir, configDir)).Run()
	return err
}
