package main

import "testing"

// Ce fichier regroupe des tests automatisés pour vérifier la logique du jeu
// depuis le terminal, sans passer par les menus interactifs (qui attendent une
// saisie clavier). Pour les lancer :
//
//	cd src
//	go test ./...        # résumé
//	go test -v ./...      # détail de chaque test

func newTestCharacter() Character {
	return initCharacter("Testeur", "Équilibré", 1, 100, 50, []string{})
}

func TestInitCharacter(t *testing.T) {
	c := initCharacter("Testeur", "Robuste", 3, 120, 60, []string{"Cuir Synthétique"})

	if c.Name != "Testeur" || c.Class != "Robuste" {
		t.Fatalf("nom/classe incorrects : got %q/%q", c.Name, c.Class)
	}
	if c.Level != 3 || c.MaxHP != 120 || c.CurrentHP != 60 {
		t.Errorf("stats incorrectes : niveau=%d maxHP=%d pv=%d", c.Level, c.MaxHP, c.CurrentHP)
	}
	if c.Gold != 100 {
		t.Errorf("l'or de départ devrait être 100, got %d", c.Gold)
	}
	if c.MaxSlots != 10 {
		t.Errorf("la capacité d'inventaire de départ devrait être 10, got %d", c.MaxSlots)
	}
	if len(c.Skills) != 1 || c.Skills[0] != "Coup de Poing" {
		t.Errorf("le sort de départ devrait être 'Coup de Poing', got %v", c.Skills)
	}
	if c.Initiative < 5 || c.Initiative > 14 {
		t.Errorf("initiative hors intervalle attendu [5,14] : got %d", c.Initiative)
	}
}

func TestAddInventoryRespectsLimit(t *testing.T) {
	c := newTestCharacter()
	c.MaxSlots = 2

	if !addInventory(&c, "Eau d'Amande") {
		t.Fatal("le premier ajout aurait dû réussir")
	}
	if !addInventory(&c, "Eau Croupie") {
		t.Fatal("le deuxième ajout aurait dû réussir")
	}
	if addInventory(&c, "Stimulant") {
		t.Error("l'ajout aurait dû échouer : inventaire plein")
	}
	if len(c.Inventory) != 2 {
		t.Errorf("inventaire attendu de taille 2, got %d", len(c.Inventory))
	}
}

func TestRemoveInventory(t *testing.T) {
	c := newTestCharacter()
	addInventory(&c, "Eau d'Amande")

	if !removeInventory(&c, "Eau d'Amande") {
		t.Fatal("la suppression aurait dû réussir")
	}
	if len(c.Inventory) != 0 {
		t.Errorf("inventaire attendu vide, got %v", c.Inventory)
	}
	if removeInventory(&c, "Eau d'Amande") {
		t.Error("la suppression d'un objet absent aurait dû échouer")
	}
}

func TestTakePotHealsAndCaps(t *testing.T) {
	c := newTestCharacter()
	c.CurrentHP = 10
	c.MaxHP = 100
	addInventory(&c, "Eau d'Amande")

	takePot(&c)
	if c.CurrentHP != 60 {
		t.Errorf("PV attendus 60 (10+50), got %d", c.CurrentHP)
	}
	if countItem(&c, "Eau d'Amande") != 0 {
		t.Error("l'Eau d'Amande aurait dû être consommée")
	}

	// Vérifie le plafond au PV max.
	c.CurrentHP = 90
	addInventory(&c, "Eau d'Amande")
	takePot(&c)
	if c.CurrentHP != 100 {
		t.Errorf("PV attendus plafonnés à 100, got %d", c.CurrentHP)
	}
}

func TestPoisonPotDamagesOverTime(t *testing.T) {
	c := newTestCharacter()
	c.CurrentHP = 100
	c.MaxHP = 100
	addInventory(&c, "Eau Croupie")

	poisonPot(&c) // 3 x 10 dégâts (~3 secondes réelles à cause des time.Sleep)

	if c.CurrentHP != 70 {
		t.Errorf("PV attendus 70 (100-30), got %d", c.CurrentHP)
	}
	if countItem(&c, "Eau Croupie") != 0 {
		t.Error("l'Eau Croupie aurait dû être consommée")
	}
}

func TestIsDeadRevivesAtHalfHP(t *testing.T) {
	c := newTestCharacter()
	c.MaxHP = 100
	c.CurrentHP = 0

	if !isDead(&c) {
		t.Fatal("isDead aurait dû détecter la mort à 0 PV")
	}
	if c.CurrentHP != 50 {
		t.Errorf("PV attendus 50 après réanimation, got %d", c.CurrentHP)
	}

	c.CurrentHP = 5
	if isDead(&c) {
		t.Error("isDead ne devrait pas se déclencher au-dessus de 0 PV")
	}
}

func TestGainExperienceLevelsUp(t *testing.T) {
	c := newTestCharacter()
	c.Level = 1
	c.MaxHP = 100
	c.CurrentHP = 50
	c.Experience = 90
	c.ExperienceMax = 100

	gainExperience(&c, 20)

	if c.Level != 2 {
		t.Errorf("niveau attendu 2, got %d", c.Level)
	}
	if c.Experience != 10 {
		t.Errorf("expérience restante attendue 10 (90+20-100), got %d", c.Experience)
	}
	if c.MaxHP != 110 {
		t.Errorf("PV max attendus 110 (+10 bonus de niveau), got %d", c.MaxHP)
	}
	if c.CurrentHP != c.MaxHP {
		t.Errorf("les PV actuels devraient être restaurés au max après level up, got %d/%d", c.CurrentHP, c.MaxHP)
	}
	if c.ExperienceMax != 150 {
		t.Errorf("le seuil d'expérience attendu est 150 (100*1.5), got %d", c.ExperienceMax)
	}
}

func TestEquipAppliesBonusAndSwaps(t *testing.T) {
	c := newTestCharacter()
	c.MaxHP = 100
	c.MaxSlots = 10
	addInventory(&c, "Casque de Chantier")
	addInventory(&c, "Casque de Chantier") // un deuxième exemplaire pour tester l'échange

	equip(&c, "Casque de Chantier")
	if c.Equipment.Tete != "Casque de Chantier" {
		t.Fatalf("le casque aurait dû être équipé, got %q", c.Equipment.Tete)
	}
	if c.MaxHP != 110 {
		t.Errorf("PV max attendus 110 (+10), got %d", c.MaxHP)
	}

	// Ré-équiper le même emplacement doit remettre l'ancien objet dans l'inventaire
	// et ne pas cumuler le bonus de PV.
	equip(&c, "Casque de Chantier")
	if c.MaxHP != 110 {
		t.Errorf("PV max ne devraient pas cumuler après un échange, got %d", c.MaxHP)
	}
	if countItem(&c, "Casque de Chantier") != 1 {
		t.Errorf("l'ancien casque aurait dû revenir dans l'inventaire, inventaire=%v", c.Inventory)
	}
}

func TestCraftEquipmentChecksResourcesAndGold(t *testing.T) {
	c := newTestCharacter()
	c.Gold = 100

	// Sans matériaux : la fabrication doit échouer sans rien consommer.
	craftEquipment(&c, "Casque de Chantier")
	if len(c.Inventory) != 0 || c.Gold != 100 {
		t.Fatalf("aucune ressource ne devrait être consommée sans matériaux : inventaire=%v or=%d", c.Inventory, c.Gold)
	}

	addInventory(&c, "Plume d'Oiseau-Cri")
	addInventory(&c, "Cuir Synthétique")
	craftEquipment(&c, "Casque de Chantier")

	if c.Gold != 95 {
		t.Errorf("or attendu 95 (100-5), got %d", c.Gold)
	}
	if countItem(&c, "Casque de Chantier") != 1 {
		t.Errorf("le casque fabriqué devrait être dans l'inventaire, got %v", c.Inventory)
	}
	if countItem(&c, "Plume d'Oiseau-Cri") != 0 || countItem(&c, "Cuir Synthétique") != 0 {
		t.Errorf("les matériaux auraient dû être consommés, got %v", c.Inventory)
	}
}

func TestCraftEquipmentFailsWithoutEnoughGold(t *testing.T) {
	c := newTestCharacter()
	c.Gold = 0
	addInventory(&c, "Plume d'Oiseau-Cri")
	addInventory(&c, "Cuir Synthétique")

	craftEquipment(&c, "Casque de Chantier")

	if countItem(&c, "Casque de Chantier") != 0 {
		t.Error("la fabrication n'aurait pas dû aboutir sans jetons")
	}
	if countItem(&c, "Plume d'Oiseau-Cri") != 1 || countItem(&c, "Cuir Synthétique") != 1 {
		t.Error("les matériaux n'auraient pas dû être consommés si la fabrication échoue")
	}
}

func TestBuyItem(t *testing.T) {
	c := newTestCharacter()
	c.Gold = 10

	buyItem(&c, "Eau d'Amande") // coûte 3
	if c.Gold != 7 {
		t.Errorf("or attendu 7 (10-3), got %d", c.Gold)
	}
	if countItem(&c, "Eau d'Amande") != 1 {
		t.Error("l'Eau d'Amande achetée devrait être dans l'inventaire")
	}

	c.Gold = 1
	buyItem(&c, "Manuel : Cocktail Molotov") // coûte 25, trop cher
	if c.Gold != 1 {
		t.Errorf("l'or ne devrait pas bouger sur un achat refusé, got %d", c.Gold)
	}
	if countItem(&c, "Manuel : Cocktail Molotov") != 0 {
		t.Error("l'objet ne devrait pas être ajouté si l'achat est refusé")
	}
}

func TestUpgradeInventorySlotLimitedToThree(t *testing.T) {
	c := newTestCharacter()
	for i := 0; i < 4; i++ {
		addInventory(&c, "Sac à Dos Amélioré")
	}

	for i := 0; i < 4; i++ {
		upgradeInventorySlot(&c)
	}

	if c.SlotUpgrades != 3 {
		t.Errorf("le nombre d'améliorations devrait être plafonné à 3, got %d", c.SlotUpgrades)
	}
	if c.MaxSlots != 40 {
		t.Errorf("capacité attendue 40 (10+3*10), got %d", c.MaxSlots)
	}
	if countItem(&c, "Sac à Dos Amélioré") != 1 {
		t.Errorf("le 4e sac n'aurait pas dû être consommé (limite atteinte), got %d restant(s)", countItem(&c, "Sac à Dos Amélioré"))
	}
}

func TestSpellBookLearnsOnlyOnce(t *testing.T) {
	c := newTestCharacter()
	addInventory(&c, "Manuel : Cocktail Molotov")
	addInventory(&c, "Manuel : Cocktail Molotov")

	spellBook(&c)
	if len(c.Skills) != 2 || c.Skills[1] != "Cocktail Molotov" {
		t.Fatalf("le sort Cocktail Molotov aurait dû être appris, sorts=%v", c.Skills)
	}

	spellBook(&c) // deuxième manuel : le sort est déjà connu
	if len(c.Skills) != 2 {
		t.Errorf("le sort ne devrait pas être appris deux fois, sorts=%v", c.Skills)
	}
	if countItem(&c, "Manuel : Cocktail Molotov") != 1 {
		t.Error("le deuxième manuel n'aurait pas dû être consommé puisque le sort est déjà connu")
	}
}

func TestBasicAttackDoesNotGoBelowZero(t *testing.T) {
	m := initGoblin()
	m.CurrentHP = 3

	c := newTestCharacter()
	basicAttack(&c, &m)

	if m.CurrentHP != 0 {
		t.Errorf("les PV du monstre ne devraient pas être négatifs, got %d", m.CurrentHP)
	}
}

func TestGoblinPatternDoublesDamageEveryThirdTurn(t *testing.T) {
	m := initGoblin() // Attack = 5
	c := newTestCharacter()
	c.MaxHP = 1000
	c.CurrentHP = 1000

	goblinPattern(&m, &c, 1) // tour normal : 100%
	if c.CurrentHP != 995 {
		t.Errorf("dégâts normaux attendus 5, PV=%d", c.CurrentHP)
	}

	goblinPattern(&m, &c, 3) // tour multiple de 3 : 200%
	if c.CurrentHP != 985 {
		t.Errorf("dégâts doublés attendus 10, PV=%d", c.CurrentHP)
	}
}
