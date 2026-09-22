package main

import "testing"

// Tâche 11 : lettres uniquement, majuscule puis minuscules.
func TestFormatNom(t *testing.T) {
	cas := map[string]string{
		"louis":    "Louis",
		"LOUIS":    "Louis",
		"lOuIs42":  "Louis",
		"jean-luc": "Jeanluc",
		"élodie":   "Élodie",
		"123":      "",
	}
	for entree, attendu := range cas {
		if got := FormatNom(entree); got != attendu {
			t.Errorf("FormatNom(%q) = %q, attendu %q", entree, got, attendu)
		}
	}
}

// Tâche 11 : PV de départ = moitié des PV max, selon la classe.
func TestStatsDeClasse(t *testing.T) {
	attendus := map[string]int{"Humain": 100, "Elfe": 80, "Nain": 120}
	for _, c := range Classes {
		h := NewCharacter("Test", c)
		if h.PVMax != attendus[c.Base] {
			t.Errorf("%s : PV max %d, attendu %d", c.Nom, h.PVMax, attendus[c.Base])
		}
		if h.PV != h.PVMax/2 {
			t.Errorf("%s : PV de départ %d, attendu %d", c.Nom, h.PV, h.PVMax/2)
		}
		if h.Niveau != 1 {
			t.Errorf("%s : niveau %d, attendu 1", c.Nom, h.Niveau)
		}
	}
}

// Tâche 12 : impossible de dépasser 10 objets.
func TestLimiteInventaire(t *testing.T) {
	h := NewCharacter("Test", Classes[0])
	for len(h.Inventaire) < InventaireBase {
		if !h.AddItem("Barre de fer") {
			t.Fatal("ajout refusé avant d'atteindre la limite")
		}
	}
	if h.AddItem("Barre de fer") {
		t.Error("un 11e objet a été accepté")
	}
	if len(h.Inventaire) != 10 {
		t.Errorf("%d objets, attendu 10", len(h.Inventaire))
	}
}

// Tâche 18 : +10 places, 3 fois au maximum.
func TestAmeliorationInventaire(t *testing.T) {
	h := NewCharacter("Test", Classes[0])
	for i := 1; i <= 3; i++ {
		if !h.UpgradeInventory() {
			t.Fatalf("amélioration %d refusée", i)
		}
	}
	if h.UpgradeInventory() {
		t.Error("une 4e amélioration a été acceptée")
	}
	if h.Capacite != 40 {
		t.Errorf("capacité %d, attendu 40", h.Capacite)
	}
}

// Tâche 5 : l'Almond Water rend 50 PV sans dépasser le maximum.
func TestAlmondWater(t *testing.T) {
	h := NewCharacter("Test", Classes[1]) // 80 PV max, départ à 40
	h.UseItem(0)
	if h.PV != 80 {
		t.Errorf("PV = %d, attendu 80 (plafonné)", h.PV)
	}
	if len(h.Inventaire) != 2 {
		t.Errorf("%d bouteilles restantes, attendu 2", len(h.Inventaire))
	}
	h.UseItem(0) // PV déjà au max : la bouteille n'est pas gaspillée
	if len(h.Inventaire) != 2 {
		t.Error("une bouteille a été consommée alors que les PV étaient pleins")
	}
}

// Tâche : argent de départ.
func TestStartingCoins(t *testing.T) {
	h := NewCharacter("Test", Classes[0])
	if h.Pieces != StartingCoins {
		t.Errorf("pièces de départ = %d, attendu %d", h.Pieces, StartingCoins)
	}
}

// Tâche : mort puis résurrection à 50 % des PV.
func TestDieResurrectsAtHalfPV(t *testing.T) {
	h := NewCharacter("Test", Classes[0])
	h.PV = 0
	h.Die()
	if h.PV != h.PVMax/2 {
		t.Errorf("PV après résurrection = %d, attendu %d", h.PV, h.PVMax/2)
	}
	if h.PV <= 0 {
		t.Error("le personnage ne devrait pas ressusciter avec 0 PV")
	}
}

// Mission bonus « expérience » : gagner assez d'XP fait monter de
// niveau et augmente les PV/énergie max.
func TestAddXPLevelsUp(t *testing.T) {
	h := NewCharacter("Test", Classes[0])
	maxAvant := h.PVMax
	gained := h.AddXP(h.XPMax)
	if gained != 1 {
		t.Fatalf("niveaux gagnés = %d, attendu 1", gained)
	}
	if h.Niveau != 2 {
		t.Errorf("niveau = %d, attendu 2", h.Niveau)
	}
	if h.PVMax != maxAvant+NiveauBonusPVParPalier {
		t.Errorf("PVMax = %d, attendu %d", h.PVMax, maxAvant+NiveauBonusPVParPalier)
	}
}
