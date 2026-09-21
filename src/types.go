package main

// Equipment représente les emplacements d'équipement portés par le personnage.
type Equipment struct {
	Tete  string
	Torse string
	Pieds string
}

// Character représente le survivant incarné par le joueur dans les Backrooms.
type Character struct {
	Name          string
	Class         string
	Level         int
	MaxHP         int
	CurrentHP     int
	Inventory     []string
	MaxSlots      int
	SlotUpgrades  int
	Skills        []string
	Gold          int
	Equipment     Equipment
	Initiative    int
	Experience    int
	ExperienceMax int
	Mana          int
	MaxMana       int
	QuestProgress map[string]int
}

// Monster représente une entité errante des Backrooms affrontée en combat.
type Monster struct {
	Name       string
	MaxHP      int
	CurrentHP  int
	Attack     int
	Initiative int
	ExpReward  int
}

// ItemInfo décrit un objet consommable ou un matériau vendu par le Troqueur.
type ItemInfo struct {
	Name     string
	Price    int
	Category string // "potion_vie", "potion_poison", "potion_mana", "grimoire", "materiau", "amelioration"
}

// EquipmentInfo décrit un équipement fabricable par le Bricoleur.
type EquipmentInfo struct {
	Name    string
	Slot    string // "Tete", "Torse", "Pieds"
	HPBonus int
	Cost    int
	Recipe  map[string]int
}

// itemCatalog liste tous les objets proposés par le Troqueur.
var itemCatalog = map[string]ItemInfo{
	"Eau d'Amande":              {"Eau d'Amande", 3, "potion_vie"},
	"Eau Croupie":               {"Eau Croupie", 6, "potion_poison"},
	"Stimulant":                 {"Stimulant", 5, "potion_mana"},
	"Manuel : Cocktail Molotov": {"Manuel : Cocktail Molotov", 25, "grimoire"},
	"Lambeau de Moquette":       {"Lambeau de Moquette", 4, "materiau"},
	"Peau d'Entité":             {"Peau d'Entité", 7, "materiau"},
	"Cuir Synthétique":          {"Cuir Synthétique", 3, "materiau"},
	"Plume d'Oiseau-Cri":        {"Plume d'Oiseau-Cri", 1, "materiau"},
	"Sac à Dos Amélioré":        {"Sac à Dos Amélioré", 30, "amelioration"},
}

// shopOrder fixe l'ordre d'affichage des objets du Troqueur.
var shopOrder = []string{
	"Eau d'Amande",
	"Eau Croupie",
	"Stimulant",
	"Manuel : Cocktail Molotov",
	"Lambeau de Moquette",
	"Peau d'Entité",
	"Cuir Synthétique",
	"Plume d'Oiseau-Cri",
	"Sac à Dos Amélioré",
}

// equipmentCatalog liste les équipements fabricables par le Bricoleur.
var equipmentCatalog = map[string]EquipmentInfo{
	"Casque de Chantier": {
		Name: "Casque de Chantier", Slot: "Tete", HPBonus: 10, Cost: 5,
		Recipe: map[string]int{"Plume d'Oiseau-Cri": 1, "Cuir Synthétique": 1},
	},
	"Veste Matelassée": {
		Name: "Veste Matelassée", Slot: "Torse", HPBonus: 25, Cost: 5,
		Recipe: map[string]int{"Lambeau de Moquette": 2, "Peau d'Entité": 1},
	},
	"Bottes de Randonnée": {
		Name: "Bottes de Randonnée", Slot: "Pieds", HPBonus: 15, Cost: 5,
		Recipe: map[string]int{"Lambeau de Moquette": 1, "Cuir Synthétique": 1},
	},
}

// equipmentOrder fixe l'ordre d'affichage des équipements du Bricoleur.
var equipmentOrder = []string{"Casque de Chantier", "Veste Matelassée", "Bottes de Randonnée"}

// spellDamage associe chaque sort à ses dégâts.
var spellDamage = map[string]int{
	"Coup de Poing":    8,
	"Cocktail Molotov": 18,
}

// spellCost associe chaque sort à son coût en Adrénaline (mana).
var spellCost = map[string]int{
	"Coup de Poing":    5,
	"Cocktail Molotov": 15,
}
