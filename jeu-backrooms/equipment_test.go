package main

import "testing"

// Tâche : struct Equipment (tête, torse, pieds) avec bonus de PV max.
func TestEquipmentBonusPV(t *testing.T) {
	var e Equipment
	if e.TotalBonusPV() != 0 || e.Count() != 0 {
		t.Fatal("un équipement vide ne doit rien bonifier")
	}
	e.Set(&ItemCasqueTuyau)
	e.Set(&ItemPlastronMoquette)
	e.Set(&ItemBottesFer)
	want := ItemCasqueTuyau.BonusPV + ItemPlastronMoquette.BonusPV + ItemBottesFer.BonusPV
	if got := e.TotalBonusPV(); got != want {
		t.Errorf("bonus total = %d, attendu %d", got, want)
	}
	if e.Count() != 3 {
		t.Errorf("Count() = %d, attendu 3", e.Count())
	}
}

// Tâche : échange avec l'ancien objet — équiper une nouvelle pièce sur
// un emplacement déjà occupé renvoie l'ancienne.
func TestEquipmentExchange(t *testing.T) {
	var e Equipment
	old := e.Set(&ItemCasqueTuyau)
	if old != nil {
		t.Fatal("le premier équipement ne devrait rien renvoyer")
	}
	autre := EquipmentItem{"Casque en tôle", SlotHead, 5}
	old = e.Set(&autre)
	if old == nil || old.Nom != ItemCasqueTuyau.Nom {
		t.Fatalf("l'ancien casque aurait dû être renvoyé, obtenu %v", old)
	}
	if e.Head.Nom != autre.Nom {
		t.Error("le nouveau casque n'a pas remplacé l'ancien")
	}
}

// L'équipement recalcule PVMax et renvoie l'objet précédent à
// l'inventaire (EquipItem, player.go).
func TestCharacterEquipItemUpdatesMaxPVAndReturnsOld(t *testing.T) {
	h := NewCharacter("Test", Classes[0])
	baseMax := h.PVMax
	h.EquipItem(&ItemCasqueTuyau)
	if h.PVMax != baseMax+ItemCasqueTuyau.BonusPV {
		t.Errorf("PVMax = %d, attendu %d", h.PVMax, baseMax+ItemCasqueTuyau.BonusPV)
	}
	autre := EquipmentItem{"Vieux casque", SlotHead, 3}
	h.EquipItem(&autre)
	if h.PVMax != baseMax+autre.BonusPV {
		t.Errorf("PVMax après échange = %d, attendu %d", h.PVMax, baseMax+autre.BonusPV)
	}
	if countItem(h.Inventaire, ItemCasqueTuyau.Nom) != 1 {
		t.Error("l'ancien casque aurait dû revenir dans l'inventaire")
	}
}

// Retirer un objet d'équipement ne doit jamais faire descendre les PV
// courants sous 0, ni les PV max en dessous de la base de la classe.
func TestUnequipNeverDropsBelowBase(t *testing.T) {
	h := NewCharacter("Test", Classes[0])
	h.EquipItem(&ItemPlastronMoquette)
	h.PV = h.PVMax
	h.Equip.Chest = nil
	h.RecomputeMaxStats()
	if h.PVMax != h.BasePVMax {
		t.Errorf("PVMax après retrait = %d, attendu %d (base de classe)", h.PVMax, h.BasePVMax)
	}
	if h.PV > h.PVMax {
		t.Error("les PV courants dépassent le nouveau maximum")
	}
}
