package main

import "testing"

func TestKnowsSpellDefaultsFalse(t *testing.T) {
	h := NewCharacter("Test", Classes[0])
	if h.KnowsSpell(SpellGrimoire) {
		t.Error("aucun sort ne devrait être connu à la création")
	}
}

// Tâche : livre de sort appris une seule fois.
func TestLearnSpellOnlyOnce(t *testing.T) {
	h := NewCharacter("Test", Classes[0])
	if err := h.LearnSpell(SpellGrimoire); err != nil {
		t.Fatalf("premier apprentissage refusé : %v", err)
	}
	if !h.KnowsSpell(SpellGrimoire) {
		t.Fatal("le sort devrait être connu")
	}
	if err := h.LearnSpell(SpellGrimoire); err != ErrSpellAlreadyKnown {
		t.Errorf("erreur = %v, attendu ErrSpellAlreadyKnown", err)
	}
}

func TestAttackLockedOnlyForSpellAttacks(t *testing.T) {
	h := NewCharacter("Test", Classes[0])
	for _, a := range Attacks {
		locked := h.AttackLocked(a)
		wantLocked := a.RequiresSpell != ""
		if locked != wantLocked {
			t.Errorf("%s : verrouillé=%v, attendu %v", a.Nom, locked, wantLocked)
		}
	}
}
