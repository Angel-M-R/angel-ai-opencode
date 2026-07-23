// Package tui implements the step-by-step installation wizard: one step per
// asset category with checkboxes, an extras step for standalone TUI toggles,
// a confirmation step with the install plan, and a final report.
package tui

import (
	"fmt"
	"slices"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"angel-ai-opencode/internal/assets"
	"angel-ai-opencode/internal/catalog"
	"angel-ai-opencode/internal/install"
)

var (
	titleStyle             = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("212"))
	stepStyle              = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	cursorStyle            = lipgloss.NewStyle().Foreground(lipgloss.Color("212"))
	checkedStyle           = lipgloss.NewStyle().Foreground(lipgloss.Color("42"))
	helpStyle              = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	noChangeActionStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	successActionStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("42"))
	updateActionStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("220"))
	replacementActionStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("212"))
	errorActionStyle       = lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
	selectedCount          = lipgloss.NewStyle().Foreground(lipgloss.Color("110"))
)

var spinnerFrames = [...]string{"|", "/", "-", "\\"}

const spinnerInterval = 100 * time.Millisecond

type installerActionPrefix struct {
	label string
	style lipgloss.Style
}

var installerActionPrefixes = []installerActionPrefix{
	{label: "SIN CAMBIOS", style: noChangeActionStyle},
	{label: "sin cambios", style: noChangeActionStyle},
	{label: "CREAR", style: successActionStyle},
	{label: "INSTALAR", style: successActionStyle},
	{label: "creado", style: successActionStyle},
	{label: "instalado", style: successActionStyle},
	{label: "ACTUALIZAR", style: updateActionStyle},
	{label: "actualizado", style: updateActionStyle},
	{label: "REEMPLAZAR", style: replacementActionStyle},
	{label: "backup", style: replacementActionStyle},
	{label: "ERROR", style: errorActionStyle},
}

func recognizedInstallerActionPrefix(line string) (installerActionPrefix, bool) {
	for _, prefix := range installerActionPrefixes {
		if strings.HasPrefix(line, prefix.label) &&
			(len(line) == len(prefix.label) || line[len(prefix.label)] == ' ') {
			return prefix, true
		}
	}
	return installerActionPrefix{}, false
}

type phase int

const (
	selecting phase = iota
	extrasPhase
	analyzing
	confirming
	installing
	finished
)

type spinnerTickMsg struct{}

type plannedMsg struct {
	plan []string
}

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

	assets    assets.Source
	configDir string

	phase  phase
	step   int
	cursor int
	report []string
	err    error

	width              int
	height             int
	confirmationPlan   []string
	confirmationOffset int
	resultOffset       int
	spinnerFrame       int
	installAfterPlan   bool
	asyncFlow          bool
}

// New builds the wizard with every catalog item preselected and extras set to
// their descriptor defaults.
func New(categories []catalog.Category, assetSource assets.Source, configDir string) Model {
	selected := make([][]bool, len(categories))
	for i, category := range categories {
		selected[i] = make([]bool, len(category.Items))
		for j := range selected[i] {
			selected[i][j] = true
		}
	}
	extras := install.ExtraOptions
	extraSelected := make([]bool, len(extras))
	for i, extra := range extras {
		extraSelected[i] = extra.DefaultSelected
	}
	return Model{
		categories:    categories,
		selected:      selected,
		extras:        extras,
		extraSelected: extraSelected,
		assets:        assetSource,
		configDir:     configDir,
		asyncFlow:     true,
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
	case spinnerTickMsg:
		if m.phase != analyzing && m.phase != installing {
			return m, nil
		}
		m.spinnerFrame = (m.spinnerFrame + 1) % len(spinnerFrames)
		return m, spinnerTick()
	case plannedMsg:
		if m.phase != analyzing {
			return m, nil
		}
		if m.installAfterPlan && slices.Equal(msg.plan, m.confirmationPlan) {
			return m.startInstallation()
		}
		m.confirmationPlan = msg.plan
		m.confirmationOffset = 0
		m.phase = confirming
		return m, nil
	case installedMsg:
		m.report = msg.report
		m.err = msg.err
		m.phase = finished
		m.resultOffset = 0
		return m, nil
	case tea.WindowSizeMsg:
		oldWidth := m.width
		m.width = msg.Width
		m.height = msg.Height
		switch m.phase {
		case confirming:
			rows := listVisualRows(m.confirmationPlan, m.width)
			height := m.confirmationListHeight(len(rows))
			if oldWidth != m.width {
				m.confirmationOffset = resizeListOffset(m.confirmationPlan, m.confirmationOffset, oldWidth, m.width, height)
			} else {
				m.confirmationOffset = clampListOffset(m.confirmationOffset, len(rows), height)
			}
		case finished:
			rows := listVisualRows(m.report, m.width)
			height := m.resultListHeight(len(rows))
			if oldWidth != m.width {
				m.resultOffset = resizeListOffset(m.report, m.resultOffset, oldWidth, m.width, height)
			} else {
				m.resultOffset = clampListOffset(m.resultOffset, len(rows), height)
			}
		}
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
			return m.updateFinished(key)
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
			return m, m.proceedToConfirmation()
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
		return m, m.proceedToConfirmation()
	}
	return m, nil
}

func spinnerTick() tea.Cmd {
	return tea.Tick(spinnerInterval, func(time.Time) tea.Msg {
		return spinnerTickMsg{}
	})
}

func (m Model) planCmd() tea.Msg {
	return plannedMsg{plan: m.installationPlan()}
}

func (m *Model) proceedToConfirmation() tea.Cmd {
	if !m.asyncFlow {
		m.enterConfirmation()
		return nil
	}
	return m.startAnalysis(false)
}

func (m *Model) startAnalysis(installAfterPlan bool) tea.Cmd {
	m.phase = analyzing
	m.spinnerFrame = 0
	m.installAfterPlan = installAfterPlan
	return tea.Batch(m.planCmd, spinnerTick())
}

func (m *Model) enterConfirmation() {
	m.asyncFlow = false
	m.phase = confirming
	m.confirmationOffset = 0
	m.confirmationPlan = m.installationPlan()
}

func (m Model) installationPlan() []string {
	plan, err := install.PlanInstallation(install.InstallationRequest{
		Items: m.chosen(), Extras: m.chosenExtras(), Assets: m.assets, ConfigDir: m.configDir,
	})
	if err != nil {
		return []string{"ERROR      " + err.Error()}
	}
	return plan
}

func (m Model) confirmationListHeight(totalRows int) int {
	return availableListHeight(m.height, m.confirmationChromeHeight(), totalRows)
}

func (m Model) resultListHeight(totalRows int) int {
	return availableListHeight(m.height, m.resultChromeHeight(), totalRows)
}

func availableListHeight(terminalHeight, chromeHeight, totalRows int) int {
	if totalRows == 0 {
		return 0
	}
	if terminalHeight <= 0 {
		return totalRows
	}
	height := terminalHeight - chromeHeight
	if height < 0 {
		return 0
	}
	return height
}

func visualHeight(rendered string, width int) int {
	if rendered == "" {
		return 0
	}

	height := 0
	for len(rendered) > 0 {
		line, rest, found := strings.Cut(rendered, "\n")
		if !found {
			height += visualLineHeight(line, width)
			break
		}
		height += visualLineHeight(line, width)
		rendered = rest
	}
	return height
}

func visualLineHeight(line string, width int) int {
	if width <= 0 {
		return 1
	}
	lineWidth := lipgloss.Width(line)
	if lineWidth == 0 {
		return 1
	}
	return (lineWidth + width - 1) / width
}

func clampListOffset(offset, total, height int) int {
	maxOffset := total - height
	if maxOffset < 0 {
		maxOffset = 0
	}
	if offset < 0 {
		return 0
	}
	if offset > maxOffset {
		return maxOffset
	}
	return offset
}

func navigateListOffset(offset int, key string, total, height int) (int, bool) {
	switch key {
	case "up", "k":
		offset--
	case "down", "j":
		offset++
	case "pgup":
		offset -= height
	case "pgdown":
		offset += height
	case "home":
		offset = 0
	case "end":
		offset = total
	default:
		return offset, false
	}
	return clampListOffset(offset, total, height), true
}

func isListNavigationKey(key string) bool {
	switch key {
	case "up", "k", "down", "j", "pgup", "pgdown", "home", "end":
		return true
	default:
		return false
	}
}

func (m Model) updateConfirming(key string) (tea.Model, tea.Cmd) {
	if isListNavigationKey(key) {
		rows := listVisualRows(m.confirmationPlan, m.width)
		offset, _ := navigateListOffset(
			m.confirmationOffset,
			key,
			len(rows),
			m.confirmationListHeight(len(rows)),
		)
		m.confirmationOffset = offset
		return m, nil
	}

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
		if m.asyncFlow {
			return m, m.startAnalysis(true)
		}
		plan := m.installationPlan()
		if !slices.Equal(plan, m.confirmationPlan) {
			m.confirmationPlan = plan
			m.confirmationOffset = 0
			return m, nil
		}
		items := m.chosen()
		extras := m.chosenExtras()
		return m, installCmd(items, extras, m.assets, m.configDir)
	}
	return m, nil
}

func (m Model) startInstallation() (tea.Model, tea.Cmd) {
	m.phase = installing
	m.spinnerFrame = 0
	cmd := installCmd(m.chosen(), m.chosenExtras(), m.assets, m.configDir)
	return m, tea.Batch(cmd, spinnerTick())
}

func installCmd(items []catalog.Item, extras map[string]bool, assetSource assets.Source, configDir string) tea.Cmd {
	return func() tea.Msg {
		report, err := install.ApplyInstallation(install.InstallationRequest{
			Items: items, Extras: extras, Assets: assetSource, ConfigDir: configDir,
		})
		return installedMsg{report: report, err: err}
	}
}

func (m Model) updateFinished(key string) (tea.Model, tea.Cmd) {
	rows := listVisualRows(m.report, m.width)
	if offset, handled := navigateListOffset(
		m.resultOffset,
		key,
		len(rows),
		m.resultListHeight(len(rows)),
	); handled {
		m.resultOffset = offset
		return m, nil
	}
	if key == "enter" || key == "esc" {
		return m, tea.Quit
	}
	return m, nil
}

func (m Model) View() string {
	var b strings.Builder
	b.WriteString(wizardHeader())

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
		b.WriteString("  " + titleStyle.Render("Integraciones y extras") + "\n\n")
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
	case analyzing:
		b.WriteString(m.loadingView("Analizando archivos…"))
	case confirming:
		b.WriteString(m.confirmationHeader())
		rows := listVisualRows(m.confirmationPlan, m.width)
		visible, start, end := orderedListWindow(rows, m.confirmationOffset, m.confirmationListHeight(len(rows)))
		for _, row := range visible {
			b.WriteString(row + "\n")
		}
		b.WriteString(confirmationFooter(listRangeFeedback(start, end, len(m.confirmationPlan))))
	case installing:
		b.WriteString(m.loadingView("Instalando…"))
	case finished:
		b.WriteString(m.resultHeader())
		rows := listVisualRows(m.report, m.width)
		visible, start, end := orderedListWindow(rows, m.resultOffset, m.resultListHeight(len(rows)))
		for _, row := range visible {
			b.WriteString(row + "\n")
		}
		b.WriteString(resultFooter(listRangeFeedback(start, end, len(m.report))))
	}
	b.WriteString("\n")
	return b.String()
}

func (m Model) loadingView(message string) string {
	frame := spinnerFrames[m.spinnerFrame%len(spinnerFrames)]
	return cursorStyle.Render(frame) + " " + titleStyle.Render(message) + "\n\n" +
		helpStyle.Render("q salir")
}

func wizardHeader() string {
	return titleStyle.Render("Angel AI — instalador de opencode") + "\n\n"
}

func (m Model) confirmationHeader() string {
	var b strings.Builder
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
		b.WriteString(fmt.Sprintf("  Integraciones y extras: %s\n", selectedCount.Render(fmt.Sprintf("%d/%d", count, len(m.extras)))))
	}
	b.WriteString("\n" + stepStyle.Render(fmt.Sprintf("%d acciones sobre %s", len(m.confirmationPlan), m.configDir)) + "\n")
	return b.String()
}

func confirmationFooter(feedback string) string {
	return helpStyle.Render(feedback) + "\n\n" +
		helpStyle.Render("↑/↓ · pgup/pgdn · home/end · enter instalar · ← volver · q salir")
}

func (m Model) resultHeader() string {
	if m.err != nil {
		return errorActionStyle.Render("Error:") + " " + m.err.Error() + "\n\n"
	}
	return successActionStyle.Render("Instalación completada") + "\n\n"
}

func resultFooter(feedback string) string {
	return helpStyle.Render(feedback) + "\n\n" +
		helpStyle.Render("↑/↓ · pgup/pgdn · home/end · enter/esc/q salir")
}

func (m Model) confirmationChromeHeight() int {
	feedback := listRangeFeedback(max(0, len(m.confirmationPlan)-1), len(m.confirmationPlan), len(m.confirmationPlan))
	return visualHeight(wizardHeader()+m.confirmationHeader()+confirmationFooter(feedback)+"\n", m.width)
}

func (m Model) resultChromeHeight() int {
	feedback := listRangeFeedback(max(0, len(m.report)-1), len(m.report), len(m.report))
	return visualHeight(wizardHeader()+m.resultHeader()+resultFooter(feedback)+"\n", m.width)
}

type listVisualRow struct {
	text          string
	entry         int
	labelStart    int
	labelEnd      int
	labelStyle    lipgloss.Style
	hasLabelStyle bool
}

func (row listVisualRow) render() string {
	if !row.hasLabelStyle {
		return row.text
	}
	return row.text[:row.labelStart] +
		row.labelStyle.Render(row.text[row.labelStart:row.labelEnd]) +
		row.text[row.labelEnd:]
}

func listVisualRows(lines []string, width int) []listVisualRow {
	rows := make([]listVisualRow, 0, len(lines))
	for entry, line := range lines {
		indentedLine := "  " + line
		prefix, recognized := recognizedInstallerActionPrefix(line)
		labelStart := len("  ")
		labelEnd := labelStart + len(prefix.label)
		rowStart := 0
		for _, text := range wrapVisualLine(indentedLine, width) {
			row := listVisualRow{text: text, entry: entry}
			rowEnd := rowStart + len(text)
			spanStart := max(labelStart, rowStart)
			spanEnd := min(labelEnd, rowEnd)
			if recognized && spanStart < spanEnd {
				row.labelStart = spanStart - rowStart
				row.labelEnd = spanEnd - rowStart
				row.labelStyle = prefix.style
				row.hasLabelStyle = true
			}
			rows = append(rows, row)
			rowStart = rowEnd
		}
	}
	return rows
}

func resizeListOffset(lines []string, offset, oldWidth, newWidth, newHeight int) int {
	oldRows := listVisualRows(lines, oldWidth)
	if len(oldRows) == 0 {
		return 0
	}
	topEntry := oldRows[clampListOffset(offset, len(oldRows), 1)].entry

	newRows := listVisualRows(lines, newWidth)
	newOffset := 0
	for newOffset < len(newRows) && newRows[newOffset].entry < topEntry {
		newOffset++
	}
	return clampListOffset(newOffset, len(newRows), newHeight)
}

func wrapVisualLine(line string, width int) []string {
	if width <= 0 || lipgloss.Width(line) <= width {
		return []string{line}
	}

	var rows []string
	var row strings.Builder
	rowWidth := 0
	for _, character := range line {
		characterWidth := lipgloss.Width(string(character))
		if rowWidth > 0 && rowWidth+characterWidth > width {
			rows = append(rows, row.String())
			row.Reset()
			rowWidth = 0
		}
		row.WriteRune(character)
		rowWidth += characterWidth
	}
	rows = append(rows, row.String())
	return rows
}

func orderedListWindow(rows []listVisualRow, offset, height int) ([]string, int, int) {
	if len(rows) == 0 || height == 0 {
		return nil, 0, 0
	}
	offset = clampListOffset(offset, len(rows), height)
	end := offset + height
	if end > len(rows) {
		end = len(rows)
	}
	visible := make([]string, end-offset)
	for index, row := range rows[offset:end] {
		visible[index] = row.render()
	}
	return visible, rows[offset].entry, rows[end-1].entry + 1
}

func listRangeFeedback(start, end, total int) string {
	if total == 0 {
		return "Mostrando 0 de 0"
	}
	if end == 0 {
		return fmt.Sprintf("Mostrando 0 de %d", total)
	}
	return fmt.Sprintf("Mostrando %d-%d de %d", start+1, end, total)
}

// Run starts the wizard.
func Run(categories []catalog.Category, assetSource assets.Source, configDir string) error {
	_, err := tea.NewProgram(New(categories, assetSource, configDir)).Run()
	return err
}
