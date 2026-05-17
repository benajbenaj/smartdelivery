package patch

import (
	"errors"
	"testing"
)

type rulePatch struct {
	Name     *string
	Priority *int
	Enabled  *bool
}

type ruleTarget struct {
	Name     string
	Priority int
	Enabled  bool
}

func TestApplyCopiesOnlyNonNilPatchFields(t *testing.T) {
	name := "Eco first"
	enabled := false
	target := ruleTarget{
		Name:     "Old name",
		Priority: 10,
		Enabled:  true,
	}

	err := Apply(&target, rulePatch{
		Name:    &name,
		Enabled: &enabled,
	})
	if err != nil {
		t.Fatalf("Apply() error = %v", err)
	}

	if target.Name != "Eco first" {
		t.Fatalf("Name = %q, want Eco first", target.Name)
	}
	if target.Priority != 10 {
		t.Fatalf("Priority = %d, want unchanged 10", target.Priority)
	}
	if target.Enabled {
		t.Fatalf("Enabled = true, want false")
	}
}

func TestApplyAcceptsPointerPatch(t *testing.T) {
	priority := 25
	target := ruleTarget{Priority: 10}

	err := Apply(&target, &rulePatch{Priority: &priority})
	if err != nil {
		t.Fatalf("Apply() error = %v", err)
	}

	if target.Priority != 25 {
		t.Fatalf("Priority = %d, want 25", target.Priority)
	}
}

func TestApplyRejectsInvalidTarget(t *testing.T) {
	err := Apply(ruleTarget{}, rulePatch{})
	if !errors.Is(err, ErrInvalidTarget) {
		t.Fatalf("Apply() error = %v, want ErrInvalidTarget", err)
	}
}

func TestApplyRejectsNonPointerPatchField(t *testing.T) {
	type badPatch struct {
		Name string
	}

	err := Apply(&ruleTarget{}, badPatch{})
	if !errors.Is(err, ErrInvalidPatch) {
		t.Fatalf("Apply() error = %v, want ErrInvalidPatch", err)
	}
}

func TestApplyRejectsMissingTargetField(t *testing.T) {
	value := "value"
	type badPatch struct {
		Unknown *string
	}

	err := Apply(&ruleTarget{}, badPatch{Unknown: &value})
	if !errors.Is(err, ErrInvalidPatch) {
		t.Fatalf("Apply() error = %v, want ErrInvalidPatch", err)
	}
}

func TestApplyRejectsIncompatibleType(t *testing.T) {
	value := "not an int"
	type badPatch struct {
		Priority *string
	}

	err := Apply(&ruleTarget{}, badPatch{Priority: &value})
	if !errors.Is(err, ErrInvalidPatch) {
		t.Fatalf("Apply() error = %v, want ErrInvalidPatch", err)
	}
}
