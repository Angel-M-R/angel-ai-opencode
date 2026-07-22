package tui

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"angel-ai-opencode/internal/catalog"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
)

var ansiSGRPattern = regexp.MustCompile("\\x1b\\[[0-9;]*m")

func forceANSI256(t *testing.T) {
	t.Helper()

	previous := lipgloss.ColorProfile()
	lipgloss.SetColorProfile(termenv.ANSI256)
	t.Cleanup(func() {
		lipgloss.SetColorProfile(previous)
	})
}

func stripANSI(value string) string {
	return ansiSGRPattern.ReplaceAllString(value, "")
}

func ansi256Label(label, color string) string {
	return lipgloss.NewStyle().Foreground(lipgloss.Color(color)).Render(label)
}

func renderWithColorProfile(model Model, profile termenv.Profile) string {
	lipgloss.SetColorProfile(profile)
	return model.View()
}

func syntheticEntries(prefix string, count int) []string {
	entries := make([]string, count)
	for index := range entries {
		entries[index] = fmt.Sprintf("%s-%03d", prefix, index)
	}
	return entries
}

func updateTestModel(t *testing.T, model Model, msg tea.Msg) (Model, tea.Cmd) {
	t.Helper()

	updated, cmd := model.Update(msg)
	result, ok := updated.(Model)
	if !ok {
		t.Fatalf("updated model type = %T, want tui.Model", updated)
	}
	return result, cmd
}

func visibleEntryIndexes(t *testing.T, view string, entries []string) []int {
	t.Helper()

	view = strings.ReplaceAll(view, "\n", "")
	indexes := make([]int, 0, len(entries))
	previousPosition := -1
	for index, entry := range entries {
		position := strings.Index(view, entry)
		if position < 0 {
			continue
		}
		if position <= previousPosition {
			t.Fatalf("entry %q rendered out of order in:\n%s", entry, view)
		}
		previousPosition = position
		indexes = append(indexes, index)
	}
	for index := 1; index < len(indexes); index++ {
		if indexes[index] != indexes[index-1]+1 {
			t.Fatalf("visible entry indexes = %v, want a contiguous ordered range", indexes)
		}
	}
	return indexes
}

func visualOffsetForEntry(t *testing.T, rows []listVisualRow, entry int) int {
	t.Helper()

	for offset, row := range rows {
		if row.entry == entry {
			return offset
		}
	}
	t.Fatalf("entry %d has no visual row", entry)
	return 0
}

func resizeModel(t *testing.T, model Model, height int) Model {
	t.Helper()

	return resizeModelTo(t, model, 80, height)
}

func resizeModelTo(t *testing.T, model Model, width, height int) Model {
	t.Helper()

	resized, _ := updateTestModel(t, model, tea.WindowSizeMsg{Width: width, Height: height})
	return resized
}

func pressKey(t *testing.T, model Model, key tea.KeyType) (Model, tea.Cmd) {
	t.Helper()

	return updateTestModel(t, model, tea.KeyMsg{Type: key})
}

func pressRune(t *testing.T, model Model, key rune) (Model, tea.Cmd) {
	t.Helper()

	return updateTestModel(t, model, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{key}})
}

func confirmationModel(t *testing.T, entryCount int) (Model, []string) {
	t.Helper()

	assetsDir := t.TempDir()
	configDir := t.TempDir()
	items := make([]catalog.Item, entryCount)
	expected := make([]string, entryCount)
	for index := range items {
		name := fmt.Sprintf("plan-%03d.txt", index)
		source := filepath.Join(assetsDir, name)
		if err := os.WriteFile(source, []byte(name), 0o644); err != nil {
			t.Fatalf("write synthetic asset: %v", err)
		}
		destination := filepath.Join("synthetic", name)
		items[index] = catalog.Item{
			Name:   name,
			Source: source,
			Dest:   destination,
			Kind:   catalog.CopyFile,
		}
		expected[index] = "CREAR       " + filepath.Join(configDir, destination)
	}

	model := New([]catalog.Category{{
		Name:  "synthetic",
		Title: "Synthetic",
		Items: items,
	}}, assetsDir, configDir)
	model.extras = nil
	model.extraSelected = nil
	model.enterConfirmation()
	return model, expected
}

func assertConfirmationControls(t *testing.T, view string) {
	t.Helper()

	for _, control := range []string{"enter instalar", "← volver", "q salir"} {
		if !strings.Contains(view, control) {
			t.Errorf("confirmation view missing control %q:\n%s", control, view)
		}
	}
}

func reachableEntries(t *testing.T, model Model, entries []string, inspect func(string)) []int {
	t.Helper()

	model, cmd := pressKey(t, model, tea.KeyHome)
	if cmd != nil {
		t.Fatal("home navigation requested program exit")
	}
	reached := make([]int, 0, len(entries))
	seen := make([]bool, len(entries))
	for step := 0; step <= len(entries)*20; step++ {
		view := model.View()
		if inspect != nil {
			inspect(view)
		}
		indexes := visibleEntryIndexes(t, view, entries)
		for _, index := range indexes {
			if seen[index] {
				continue
			}
			if index != len(reached) {
				t.Fatalf("newly reached entry index = %d after %v, want %d", index, reached, len(reached))
			}
			seen[index] = true
			reached = append(reached, index)
		}
		if len(reached) == len(entries) {
			return reached
		}

		next, nextCmd := pressKey(t, model, tea.KeyDown)
		if nextCmd != nil {
			t.Fatal("down navigation requested program exit")
		}
		if next.View() == view {
			return reached
		}
		model = next
	}
	return reached
}

func TestInstallerActionLabelColorMapping(t *testing.T) {
	forceANSI256(t)

	tests := []struct {
		name  string
		label string
		color string
		model Model
	}{
		{name: "confirmation no change", label: "SIN CAMBIOS", color: "241", model: Model{phase: confirming, confirmationPlan: []string{"SIN CAMBIOS  /tmp/existing"}}},
		{name: "confirmation create", label: "CREAR", color: "42", model: Model{phase: confirming, confirmationPlan: []string{"CREAR       /tmp/new"}}},
		{name: "confirmation install", label: "INSTALAR", color: "42", model: Model{phase: confirming, confirmationPlan: []string{"INSTALAR    package@latest"}}},
		{name: "confirmation update", label: "ACTUALIZAR", color: "220", model: Model{phase: confirming, confirmationPlan: []string{"ACTUALIZAR  /tmp/changed"}}},
		{name: "confirmation replace", label: "REEMPLAZAR", color: "212", model: Model{phase: confirming, confirmationPlan: []string{"REEMPLAZAR  /tmp/replaced"}}},
		{name: "confirmation error", label: "ERROR", color: "196", model: Model{phase: confirming, confirmationPlan: []string{"ERROR      planning failed"}}},
		{name: "completion no change", label: "sin cambios", color: "241", model: Model{phase: finished, report: []string{"sin cambios  /tmp/existing"}}},
		{name: "completion created", label: "creado", color: "42", model: Model{phase: finished, report: []string{"creado       /tmp/new"}}},
		{name: "completion installed", label: "instalado", color: "42", model: Model{phase: finished, report: []string{"instalado    package@latest"}}},
		{name: "completion updated", label: "actualizado", color: "220", model: Model{phase: finished, report: []string{"actualizado  /tmp/changed"}}},
		{name: "completion backup", label: "backup", color: "212", model: Model{phase: finished, report: []string{"backup       /tmp/original -> /tmp/backup"}}},
		{name: "completion status", label: "Instalación completada", color: "42", model: Model{phase: finished}},
		{name: "error status", label: "Error:", color: "196", model: Model{phase: finished, err: errors.New("apply failed")}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			view := test.model.View()
			want := ansi256Label(test.label, test.color)
			if !strings.Contains(view, want) {
				t.Errorf("view does not render label %q with ANSI-256 color %s:\n%s", test.label, test.color, view)
			}
		})
	}
}

func TestInstallerActionStylesResetAtExactLabelBoundary(t *testing.T) {
	forceANSI256(t)

	tests := []struct {
		name   string
		label  string
		color  string
		suffix string
		model  Model
	}{
		{name: "confirmation path", label: "CREAR", color: "42", suffix: "       /tmp/new", model: Model{phase: confirming, confirmationPlan: []string{"CREAR       /tmp/new"}}},
		{name: "confirmation planning detail", label: "ERROR", color: "196", suffix: "      planning failed", model: Model{phase: confirming, confirmationPlan: []string{"ERROR      planning failed"}}},
		{name: "completion package specification", label: "instalado", color: "42", suffix: "    package@latest", model: Model{phase: finished, report: []string{"instalado    package@latest"}}},
		{name: "completion backup detail", label: "backup", color: "212", suffix: "       /tmp/original -> /tmp/backup", model: Model{phase: finished, report: []string{"backup       /tmp/original -> /tmp/backup"}}},
		{name: "error status detail", label: "Error:", color: "196", suffix: " apply failed", model: Model{phase: finished, err: errors.New("apply failed")}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			view := test.model.View()
			indent := "  "
			if test.label == "Error:" {
				indent = ""
			}
			want := indent + ansi256Label(test.label, test.color) + test.suffix + "\n"
			if !strings.Contains(view, want) {
				t.Errorf("view does not reset color at the boundary after %q; want line fragment %q:\n%s", test.label, want, view)
			}
			if strings.Contains(view, ansi256Label(test.label+test.suffix[:1], test.color)) {
				t.Errorf("separator after %q inherited label color:\n%s", test.label, view)
			}
		})
	}
}

func TestUnknownInstallerActionLabelsRemainUnstyled(t *testing.T) {
	forceANSI256(t)

	tests := []struct {
		name  string
		line  string
		model Model
	}{
		{name: "confirmation unknown", line: "OMITIR      /tmp/unknown", model: Model{phase: confirming, confirmationPlan: []string{"OMITIR      /tmp/unknown"}}},
		{name: "confirmation near match", line: "CREARX      /tmp/not-create", model: Model{phase: confirming, confirmationPlan: []string{"CREARX      /tmp/not-create"}}},
		{name: "completion unknown", line: "respaldado  /tmp/unknown", model: Model{phase: finished, report: []string{"respaldado  /tmp/unknown"}}},
		{name: "completion near match", line: "actualizados /tmp/not-update", model: Model{phase: finished, report: []string{"actualizados /tmp/not-update"}}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			view := test.model.View()
			want := "  " + test.line + "\n"
			if !strings.Contains(view, want) {
				t.Errorf("unknown or near-match action line was changed or styled; want raw line %q:\n%s", want, view)
			}
		})
	}
}

func TestStrippingInstallerANSIReproducesUnstyledViewsExactly(t *testing.T) {
	forceANSI256(t)

	tests := []struct {
		name  string
		model Model
	}{
		{
			name: "confirmation",
			model: Model{phase: confirming, confirmationPlan: []string{
				"SIN CAMBIOS  /tmp/existing",
				"CREAR       /tmp/new",
				"ACTUALIZAR  /tmp/changed",
				"REEMPLAZAR  /tmp/replaced",
				"ERROR      planning failed",
			}},
		},
		{
			name: "completion with error",
			model: Model{phase: finished, err: errors.New("apply failed"), report: []string{
				"sin cambios  /tmp/existing",
				"creado       /tmp/new",
				"actualizado  /tmp/changed",
				"backup       /tmp/original -> /tmp/backup",
			}},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			styled := renderWithColorProfile(test.model, termenv.ANSI256)
			plain := renderWithColorProfile(test.model, termenv.Ascii)
			if got := stripANSI(styled); got != plain {
				t.Errorf("stripping ANSI changed installer text\ngot:\n%q\nwant:\n%q", got, plain)
			}
		})
	}
}

func assertStyledViewMatchesPlain(t *testing.T, model Model) string {
	t.Helper()

	plain := renderWithColorProfile(model, termenv.Ascii)
	styled := renderWithColorProfile(model, termenv.ANSI256)
	if got := stripANSI(styled); got != plain {
		t.Errorf("styled view changed plain-text wrapping or content\ngot:\n%q\nwant:\n%q", got, plain)
	}
	return styled
}

func assertModelRangeFeedback(t *testing.T, model Model, lines []string, view string) {
	t.Helper()

	rows := listVisualRows(lines, model.width)
	offset := model.confirmationOffset
	height := model.confirmationListHeight(len(rows))
	if model.phase == finished {
		offset = model.resultOffset
		height = model.resultListHeight(len(rows))
	}
	_, start, end := orderedListWindow(rows, offset, height)
	want := listRangeFeedback(start, end, len(lines))
	if !strings.Contains(stripANSI(view), want) {
		t.Errorf("view range feedback missing %q:\n%s", want, view)
	}
}

func TestStyledInstallerListsPreserveWrappingScrollingAndResizeGeometry(t *testing.T) {
	forceANSI256(t)

	tests := []struct {
		name   string
		phase  phase
		label  string
		color  string
		prefix string
	}{
		{name: "confirmation", phase: confirming, label: "ACTUALIZAR", color: "220", prefix: "ACTUALIZAR  "},
		{name: "completion", phase: finished, label: "actualizado", color: "220", prefix: "actualizado  "},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			const (
				narrowWidth = 24
				wideWidth   = 38
				entryCount  = 12
			)
			lines := make([]string, entryCount)
			markers := make([]string, entryCount)
			for index := range lines {
				markers[index] = fmt.Sprintf("styled-item-%03d", index)
				lines[index] = test.prefix + markers[index] + "-" + strings.Repeat(string(rune('a'+index)), 42)
			}

			model := Model{phase: test.phase, width: narrowWidth}
			if test.phase == confirming {
				model.confirmationPlan = lines
				model.height = model.confirmationChromeHeight() + 5
			} else {
				model.report = lines
				model.height = model.resultChromeHeight() + 5
			}
			model = resizeModelTo(t, model, narrowWidth, model.height)

			initialView := assertStyledViewMatchesPlain(t, model)
			if !strings.Contains(initialView, ansi256Label(test.label, test.color)) {
				t.Errorf("initial wrapped view does not contain styled action label %q:\n%s", test.label, initialView)
			}
			assertModelRangeFeedback(t, model, lines, initialView)
			initialIndexes := visibleEntryIndexes(t, stripANSI(initialView), markers)
			if len(initialIndexes) == 0 || initialIndexes[0] != 0 || len(initialIndexes) >= len(markers) {
				t.Fatalf("initial visible indexes = %v, want a partial range beginning at zero", initialIndexes)
			}

			paged, cmd := pressKey(t, model, tea.KeyPgDown)
			if cmd != nil {
				t.Fatal("page-down navigation requested program exit")
			}
			pagedView := assertStyledViewMatchesPlain(t, paged)
			assertModelRangeFeedback(t, paged, lines, pagedView)
			pagedIndexes := visibleEntryIndexes(t, stripANSI(pagedView), markers)
			if len(pagedIndexes) == 0 || pagedIndexes[len(pagedIndexes)-1] <= initialIndexes[len(initialIndexes)-1] {
				t.Fatalf("page-down visible indexes = %v, want progress after %v", pagedIndexes, initialIndexes)
			}

			oldRows := listVisualRows(lines, narrowWidth)
			oldOffset := paged.confirmationOffset
			if test.phase == finished {
				oldOffset = paged.resultOffset
			}
			oldTopEntry := oldRows[oldOffset].entry
			resized := resizeModelTo(t, paged, wideWidth, paged.height)
			newRows := listVisualRows(lines, wideWidth)
			newOffset := resized.confirmationOffset
			if test.phase == finished {
				newOffset = resized.resultOffset
			}
			if got := newRows[newOffset].entry; got != oldTopEntry {
				t.Errorf("top logical entry after styled width resize = %d, want %d", got, oldTopEntry)
			}
			resizedView := assertStyledViewMatchesPlain(t, resized)
			assertModelRangeFeedback(t, resized, lines, resizedView)
			visibleEntryIndexes(t, stripANSI(resizedView), markers)

			reached := reachableEntries(t, resized, markers, func(view string) {
				if strings.Contains(stripANSI(view), "…") {
					t.Fatalf("styled list rendered an omission summary:\n%s", view)
				}
			})
			if len(reached) != len(markers) {
				t.Fatalf("reachable styled entry indexes = %v, want all %d entries", reached, len(markers))
			}

			last, cmd := pressKey(t, resized, tea.KeyEnd)
			if cmd != nil {
				t.Fatal("end navigation requested program exit")
			}
			lastView := assertStyledViewMatchesPlain(t, last)
			assertModelRangeFeedback(t, last, lines, lastView)
			lastIndexes := visibleEntryIndexes(t, stripANSI(lastView), markers)
			if len(lastIndexes) == 0 || lastIndexes[len(lastIndexes)-1] != len(markers)-1 {
				t.Fatalf("end navigation visible indexes = %v, want final entry %d", lastIndexes, len(markers)-1)
			}
		})
	}
}

func TestFinishedListWindowSizesToTerminalAndClampsOnResize(t *testing.T) {
	entries := syntheticEntries("operation", 30)
	model := resizeModel(t, Model{phase: finished, report: entries}, 10)

	compactView := model.View()
	compactIndexes := visibleEntryIndexes(t, compactView, entries)
	if len(compactIndexes) == 0 || len(compactIndexes) >= len(entries) {
		t.Fatalf("10-line window exposed %d/%d entries, want a non-empty partial range", len(compactIndexes), len(entries))
	}
	if height := strings.Count(compactView, "\n"); height > 10 {
		t.Fatalf("rendered height = %d, want at most terminal height 10", height)
	}

	model, cmd := pressKey(t, model, tea.KeyEnd)
	if cmd != nil {
		t.Fatal("end navigation requested program exit")
	}
	bottomIndexes := visibleEntryIndexes(t, model.View(), entries)
	if len(bottomIndexes) == 0 || bottomIndexes[len(bottomIndexes)-1] != len(entries)-1 {
		t.Fatalf("end navigation exposed indexes %v, want the last entry", bottomIndexes)
	}

	model = resizeModel(t, model, 40)
	expandedIndexes := visibleEntryIndexes(t, model.View(), entries)
	if len(expandedIndexes) != len(entries) {
		t.Fatalf("expanded window exposed indexes %v, want every entry", expandedIndexes)
	}

	model = resizeModel(t, model, 10)
	clampedIndexes := visibleEntryIndexes(t, model.View(), entries)
	if len(clampedIndexes) == 0 || clampedIndexes[0] != 0 {
		t.Fatalf("offset after expanding then shrinking exposed indexes %v, want a clamp from the first entry", clampedIndexes)
	}

	model = resizeModel(t, model, 14)
	largerIndexes := visibleEntryIndexes(t, model.View(), entries)
	if len(largerIndexes) <= len(compactIndexes) {
		t.Fatalf("14-line window exposed %d entries, want more than 10-line window's %d", len(largerIndexes), len(compactIndexes))
	}
}

func TestListViewsReserveWrappedVisualRowsInNarrowTerminals(t *testing.T) {
	t.Run("confirmation", func(t *testing.T) {
		model, plan := confirmationModel(t, 6)
		const width = 24
		const height = 24
		model = resizeModelTo(t, model, width, height)

		view := model.View()
		if renderedHeight := visualHeight(view, width); renderedHeight > height {
			t.Fatalf("rendered visual height = %d, want at most %d:\n%s", renderedHeight, height, view)
		}
		assertConfirmationControls(t, view)
		if indexes := visibleEntryIndexes(t, view, plan); len(indexes) >= len(plan) {
			t.Fatalf("narrow terminal exposed %d/%d plan entries, want a wrapped partial range", len(indexes), len(plan))
		}

		model, cmd := pressKey(t, model, tea.KeyEnd)
		if cmd != nil {
			t.Fatal("end navigation requested program exit")
		}
		view = model.View()
		if renderedHeight := visualHeight(view, width); renderedHeight > height {
			t.Fatalf("end view visual height = %d, want at most %d:\n%s", renderedHeight, height, view)
		}
		assertConfirmationControls(t, view)
		if !strings.Contains(view, "Mostrando 6-6 de 6") || !strings.Contains(view, "plan-005.txt") {
			t.Fatalf("end navigation did not expose the final wrapped plan entry:\n%s", view)
		}
	})

	t.Run("completion", func(t *testing.T) {
		report := []string{
			"operation-000-" + strings.Repeat("a", 36),
			"operation-001-" + strings.Repeat("b", 36),
			"operation-002-" + strings.Repeat("c", 36),
		}
		const width = 20
		const height = 14
		model := resizeModelTo(t, Model{phase: finished, report: report}, width, height)

		view := model.View()
		if renderedHeight := visualHeight(view, width); renderedHeight > height {
			t.Fatalf("rendered visual height = %d, want at most %d:\n%s", renderedHeight, height, view)
		}
		if !strings.Contains(view, "enter/esc/q salir") {
			t.Fatalf("completion view missing exit controls:\n%s", view)
		}

		model, cmd := pressKey(t, model, tea.KeyEnd)
		if cmd != nil {
			t.Fatal("end navigation requested program exit")
		}
		view = model.View()
		if renderedHeight := visualHeight(view, width); renderedHeight > height {
			t.Fatalf("end view visual height = %d, want at most %d:\n%s", renderedHeight, height, view)
		}
		if !strings.Contains(strings.ReplaceAll(view, "\n", ""), report[len(report)-1]) {
			t.Fatalf("end navigation did not expose the final wrapped report entry:\n%s", view)
		}
	})
}

func TestWidthResizePreservesTopLogicalEntry(t *testing.T) {
	entries := make([]string, 8)
	for index := range entries {
		entries[index] = fmt.Sprintf("entry-%03d-%s", index, strings.Repeat("x", 42))
	}
	const (
		narrowWidth = 20
		wideWidth   = 40
		topEntry    = 2
	)

	t.Run("confirmation", func(t *testing.T) {
		model := Model{phase: confirming, confirmationPlan: entries, width: narrowWidth}
		model.confirmationOffset = visualOffsetForEntry(t, listVisualRows(entries, narrowWidth), topEntry)
		wideModel := model
		wideModel.width = wideWidth
		model = resizeModelTo(t, model, wideWidth, wideModel.confirmationChromeHeight()+2)

		rows := listVisualRows(entries, wideWidth)
		if got := rows[model.confirmationOffset].entry; got != topEntry {
			t.Fatalf("top confirmation entry after width resize = %d, want %d", got, topEntry)
		}
	})

	t.Run("completion", func(t *testing.T) {
		model := Model{phase: finished, report: entries, width: narrowWidth}
		model.resultOffset = visualOffsetForEntry(t, listVisualRows(entries, narrowWidth), topEntry)
		wideModel := model
		wideModel.width = wideWidth
		model = resizeModelTo(t, model, wideWidth, wideModel.resultChromeHeight()+2)

		rows := listVisualRows(entries, wideWidth)
		if got := rows[model.resultOffset].entry; got != topEntry {
			t.Fatalf("top completion entry after width resize = %d, want %d", got, topEntry)
		}
	})
}

func TestListViewsAllowZeroRowsAtChromeBoundary(t *testing.T) {
	t.Run("confirmation", func(t *testing.T) {
		model, plan := confirmationModel(t, 3)
		model.width = 80
		boundaryHeight := model.confirmationChromeHeight()
		model = resizeModelTo(t, model, 80, boundaryHeight)

		view := model.View()
		if got := visualHeight(view, 80); got > boundaryHeight {
			t.Fatalf("rendered visual height = %d, want at most boundary %d:\n%s", got, boundaryHeight, view)
		}
		if indexes := visibleEntryIndexes(t, view, plan); len(indexes) != 0 {
			t.Fatalf("boundary-height view exposed entries %v, want zero list rows", indexes)
		}
		assertConfirmationControls(t, view)
	})

	t.Run("completion", func(t *testing.T) {
		report := syntheticEntries("operation", 3)
		model := Model{phase: finished, report: report, width: 80}
		boundaryHeight := model.resultChromeHeight()
		model = resizeModelTo(t, model, 80, boundaryHeight)

		view := model.View()
		if got := visualHeight(view, 80); got > boundaryHeight {
			t.Fatalf("rendered visual height = %d, want at most boundary %d:\n%s", got, boundaryHeight, view)
		}
		if indexes := visibleEntryIndexes(t, view, report); len(indexes) != 0 {
			t.Fatalf("boundary-height view exposed entries %v, want zero list rows", indexes)
		}
		if !strings.Contains(view, "enter/esc/q salir") {
			t.Fatalf("completion view missing exit controls:\n%s", view)
		}
	})
}

func TestConfirmationPlanIsCachedUntilConfirmationIsReentered(t *testing.T) {
	model, plan := confirmationModel(t, 3)
	model = resizeModel(t, model, 24)
	if err := os.RemoveAll(model.assetsDir); err != nil {
		t.Fatalf("remove assets after planning: %v", err)
	}

	for _, key := range []tea.KeyType{tea.KeyDown, tea.KeyPgDown, tea.KeyHome} {
		model, _ = pressKey(t, model, key)
		view := strings.ReplaceAll(model.View(), "\n", "")
		if strings.Contains(view, "ERROR") || !strings.Contains(view, plan[0]) {
			t.Fatalf("confirmation navigation or render refreshed the cached plan:\n%s", model.View())
		}
	}

	model, _ = pressKey(t, model, tea.KeyLeft)
	model, _ = pressKey(t, model, tea.KeyEnter)
	if view := model.View(); !strings.Contains(view, "ERROR") {
		t.Fatalf("re-entering confirmation did not refresh the installation plan:\n%s", view)
	}
}

func TestConfirmationRevalidatesChangedPlanBeforeApplying(t *testing.T) {
	model, originalPlan := confirmationModel(t, 1)
	item := model.categories[0].Items[0]
	destination := filepath.Join(model.configDir, item.Dest)
	if err := os.MkdirAll(filepath.Dir(destination), 0o755); err != nil {
		t.Fatalf("create external destination directory: %v", err)
	}
	if err := os.WriteFile(destination, []byte("external change"), 0o644); err != nil {
		t.Fatalf("create external destination: %v", err)
	}

	model, cmd := pressKey(t, model, tea.KeyEnter)
	if cmd != nil {
		t.Fatal("stale confirmation returned an install command")
	}
	if len(model.confirmationPlan) != 1 || model.confirmationPlan[0] == originalPlan[0] {
		t.Fatalf("confirmation plan after external change = %v, want a refreshed preview", model.confirmationPlan)
	}
	if view := model.View(); !strings.Contains(view, "ACTUALIZAR  "+destination) {
		t.Fatalf("refreshed confirmation does not show the changed plan:\n%s", view)
	}

	_, cmd = pressKey(t, model, tea.KeyEnter)
	if cmd == nil {
		t.Fatal("fresh confirmation did not return an install command")
	}
}

func TestFinishedListWindowFirstPageAndLastNavigationPreservesOrder(t *testing.T) {
	entries := syntheticEntries("operation", 30)
	model := resizeModel(t, Model{phase: finished, report: entries}, 10)

	firstIndexes := visibleEntryIndexes(t, model.View(), entries)
	if len(firstIndexes) == 0 || firstIndexes[0] != 0 {
		t.Fatalf("initial indexes = %v, want the first entry", firstIndexes)
	}

	model, cmd := pressKey(t, model, tea.KeyPgDown)
	if cmd != nil {
		t.Fatal("page-down navigation requested program exit")
	}
	nextIndexes := visibleEntryIndexes(t, model.View(), entries)
	if len(nextIndexes) == 0 || nextIndexes[0] <= firstIndexes[0] {
		t.Fatalf("page-down indexes = %v, want a range after %v", nextIndexes, firstIndexes)
	}

	model, cmd = pressKey(t, model, tea.KeyEnd)
	if cmd != nil {
		t.Fatal("end navigation requested program exit")
	}
	lastIndexes := visibleEntryIndexes(t, model.View(), entries)
	if len(lastIndexes) == 0 || lastIndexes[len(lastIndexes)-1] != len(entries)-1 {
		t.Fatalf("end indexes = %v, want the last entry", lastIndexes)
	}

	model, cmd = pressKey(t, model, tea.KeyPgUp)
	if cmd != nil {
		t.Fatal("page-up navigation requested program exit")
	}
	previousIndexes := visibleEntryIndexes(t, model.View(), entries)
	if len(previousIndexes) == 0 || previousIndexes[0] >= lastIndexes[0] {
		t.Fatalf("page-up indexes = %v, want a range before %v", previousIndexes, lastIndexes)
	}

	model, cmd = pressKey(t, model, tea.KeyHome)
	if cmd != nil {
		t.Fatal("home navigation requested program exit")
	}
	homeIndexes := visibleEntryIndexes(t, model.View(), entries)
	if len(homeIndexes) == 0 || homeIndexes[0] != 0 {
		t.Fatalf("home indexes = %v, want the first entry", homeIndexes)
	}
}

func TestConfirmationViewShowsShortPlanInOrderWithoutOmission(t *testing.T) {
	model, plan := confirmationModel(t, 3)
	model = resizeModel(t, model, 24)
	view := model.View()

	indexes := visibleEntryIndexes(t, view, plan)
	if len(indexes) != len(plan) {
		t.Fatalf("visible plan indexes = %v, want every entry", indexes)
	}
	if strings.Contains(view, "…") {
		t.Fatalf("confirmation view rendered an omission summary:\n%s", view)
	}
	assertConfirmationControls(t, view)
}

func TestConfirmationViewMakesOverflowingPlanReachableInOrder(t *testing.T) {
	model, plan := confirmationModel(t, 18)
	model = resizeModel(t, model, 14)

	reached := reachableEntries(t, model, plan, func(view string) {
		if strings.Contains(view, "…") {
			t.Fatalf("confirmation view rendered an omission summary:\n%s", view)
		}
		assertConfirmationControls(t, view)
	})
	if len(reached) != len(plan) {
		t.Fatalf("reachable plan indexes = %v, want all %d entries", reached, len(plan))
	}
}

func TestCompletionViewMakesSuccessfulAndPartialErrorReportsReachable(t *testing.T) {
	tests := []struct {
		name       string
		reportSize int
		err        error
		heading    string
	}{
		{
			name:       "successful report",
			reportSize: 28,
			heading:    "Instalación completada",
		},
		{
			name:       "partial report with error",
			reportSize: 24,
			err:        errors.New("synthetic apply failure"),
			heading:    "Error: synthetic apply failure",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			report := syntheticEntries("operation", test.reportSize)
			model := resizeModel(t, Model{phase: finished, report: report, err: test.err}, 10)
			reached := reachableEntries(t, model, report, func(view string) {
				if !strings.Contains(view, test.heading) {
					t.Fatalf("completion view missing heading %q:\n%s", test.heading, view)
				}
				if strings.Contains(view, "…") {
					t.Fatalf("completion view rendered an omission summary:\n%s", view)
				}
			})
			if len(reached) != len(report) {
				t.Fatalf("reachable report indexes = %v, want all %d entries", reached, len(report))
			}
		})
	}
}

func TestCompletionViewUsesNavigationAndExplicitExitKeys(t *testing.T) {
	report := syntheticEntries("operation", 30)
	model := resizeModel(t, Model{phase: finished, report: report}, 10)
	initialView := model.View()
	if strings.Contains(initialView, "cualquier tecla") {
		t.Fatalf("completion help still advertises any-key exit:\n%s", initialView)
	}
	if !strings.Contains(initialView, "q salir") {
		t.Fatalf("completion help missing explicit quit control:\n%s", initialView)
	}

	paged, cmd := pressKey(t, model, tea.KeyPgDown)
	if cmd != nil {
		t.Fatal("page-down navigation requested program exit")
	}
	if paged.View() == initialView {
		t.Fatal("page-down navigation did not change the visible report range")
	}

	unchanged, cmd := pressRune(t, paged, 'x')
	if cmd != nil {
		t.Fatal("unadvertised key requested program exit")
	}
	if unchanged.View() != paged.View() {
		t.Fatal("unadvertised key changed the completion view")
	}

	_, cmd = pressRune(t, unchanged, 'q')
	if cmd == nil {
		t.Fatal("explicit quit key did not request program exit")
	}
}

func TestCMUXSelectionDefaultsAndExplicitSelection(t *testing.T) {
	model := New(nil, t.TempDir(), t.TempDir())
	cmuxIndex := -1
	for index, extra := range model.extras {
		if extra.Key == "cmux" {
			cmuxIndex = index
			if model.extraSelected[index] {
				t.Fatal("cmux must start unselected")
			}
			continue
		}
		if !model.extraSelected[index] {
			t.Errorf("existing extra %q must remain selected by default", extra.Key)
		}
	}
	if cmuxIndex < 0 {
		t.Fatal("cmux extra is missing")
	}

	model.phase = extrasPhase
	model.cursor = cmuxIndex
	updated, _ := model.updateExtras(" ")
	model = updated.(Model)
	if !model.chosenExtras()["cmux"] {
		t.Fatal("explicit cmux selection was not included in chosen extras")
	}
}
