package tui

import "testing"

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
