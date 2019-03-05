package variables

import (
	"testing"

	"github.com/satori/go.uuid"
)

func TestVariablesRepo_CreateVariable(t *testing.T) {
	repo := NewVariablesRepo()
	pageId, _ := uuid.FromString("5d71c23d-bef4-4ccc-bbbb-12fcf6563dc5")
	state, err := repo.CreateVariable(pageId, "var1", "1234")

	if err != nil {
		t.Fatal("unexpected error:", err)
	}

	if n := state.Name; n != "var1" {
		t.Errorf("wrong name: %s; expected: var1", n)
	}

	if f := state.Formula; f != "1234" {
		t.Errorf("wrong formula: %s; expected: 1234", f)
	}
}

func TestVariablesRepo_FindPageVariables(t *testing.T) {
	repo := NewVariablesRepo()
	pageId, _ := uuid.FromString("5d71c23d-bef4-4ccc-bbbb-12fcf6563dc5")
	_, err := repo.CreateVariable(pageId, "var1", "22")
	if err != nil {
		t.Fatal("unexpected error:", err)
	}
	_, err = repo.CreateVariable(pageId, "var2", "33")
	if err != nil {
		t.Fatal("unexpected error:", err)
	}

	states := repo.FindPageVariables(pageId)
	if n := len(states); n != 2 {
		t.Fatalf("expected 2 vars; found %d", n)
	}

	// In no particular order
	var var1, var2 *VariableState
	for _, state := range states {
		if state.Name == "var1" {
			var1 = state
		} else if state.Name == "var2" {
			var2 = state
		}
	}

	if f := var1.Formula; f != "22" {
		t.Errorf("expected formula `22`; found `%s`", f)
	}

	if f := var2.Formula; f != "33" {
		t.Errorf("expected formula `33`; found `%s`", f)
	}
}
