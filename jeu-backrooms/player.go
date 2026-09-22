package main

// ---------------------------------------------------------------
// PERSONNAGE — données uniquement, aucun affichage ici.
//
// Ce fichier suit la struct Character de l'énoncé. Si la struct du back
// (tes collègues) porte d'autres noms de champs, c'est UNIQUEMENT ici qu'il
// faut adapter : l'interface (ui.go) ne lit que ces champs-là.
// ---------------------------------------------------------------

import (
	"strconv"
	"strings"
	"unicode"
)

// Character — tâche 1 (+ énergie pour la mission bonus mana).
type Character struct {
	Nom        string
	Classe     string
	Niveau     int
	PVMax      int
	PV         int
	EnergieMax int
	Energie    int
	Inventaire []string

	Capacite      int // tâche 12 : 10 emplacements au départ
	Ameliorations int // tâche 18 : +10 emplacements, 3 fois maximum

	Pieces int       // monnaie, dans un emplacement séparé de l'inventaire
	Equip  Equipment // tâche : struct Equipment (tête, torse, pieds)

	SpellsKnown map[string]bool // sorts appris (tâche : livre de sort)

	Vitesse int // mission bonus « initiative » : agit en premier si plus rapide

	XP, XPMax int // mission bonus « expérience »

	// Statistiques de base (classe), conservées à part pour pouvoir
	// recalculer PVMax/EnergieMax quand l'équipement ou le niveau changent
	// sans jamais perdre la valeur d'origine.
	BasePVMax      int
	BaseEnergieMax int
	NiveauBonusPV  int
	NiveauBonusEN  int
}

// ClassInfo décrit une classe jouable.
// Base = la classe de l'énoncé dont elle reprend les statistiques.
type ClassInfo struct {
	Nom        string
	Base       string
	PVMax      int
	EnergieMax int
	Vitesse    int
	Desc       string
}

// Classes — tâche 11 : Humain 100 PV max, Elfe 80, Nain 120.
// L'énergie et la vitesse ne sont pas fixées par l'énoncé : ce sont nos
// choix d'équilibrage (la vitesse sert à l'initiative en combat).
var Classes = []ClassInfo{
	{"Survivant", "Humain", 100, 100, 10,
		"Équilibré. Il a tenu jusqu'ici, il tiendra encore un peu."},
	{"Ancien Résident", "Elfe", 80, 120, 14,
		"Fragile, mais il connaît les failles entre les niveaux."},
	{"Chasseur de niveaux", "Nain", 120, 80, 7,
		"Encaisse les coups des entités. Se fatigue vite."},
}

const (
	StartingCoins          = 30  // tâche : argent de départ
	XPBase                 = 100 // XP nécessaire pour passer du niveau 1 au niveau 2
	XPParNiveau            = 40  // XP supplémentaire requis à chaque niveau
	NiveauBonusPVParPalier = 10  // + PV max par niveau gagné
	NiveauBonusENParPalier = 5   // + énergie max par niveau gagné
)

const (
	InventaireBase    = 10 // tâche 12
	InventaireBonus   = 10 // tâche 18
	InventaireMaxUpgr = 3  // tâche 18
	SoinAlmondWater   = 50 // tâche 5 : la potion rend 50 PV
)

// NewCharacter — tâche 11 : niveau 1, PV actuels = 50 % des PV max.
func NewCharacter(nom string, c ClassInfo) *Character {
	ch := &Character{
		Nom:            nom,
		Classe:         c.Nom,
		Niveau:         1,
		EnergieMax:     c.EnergieMax,
		Energie:        c.EnergieMax,
		Inventaire:     []string{"Almond Water", "Almond Water", "Almond Water"},
		Capacite:       InventaireBase,
		Pieces:         StartingCoins,
		Vitesse:        c.Vitesse,
		XPMax:          XPBase,
		BasePVMax:      c.PVMax,
		BaseEnergieMax: c.EnergieMax,
	}
	ch.RecomputeMaxStats()
	ch.PV = ch.PVMax / 2
	return ch
}

// RecomputeMaxStats recalcule PVMax/EnergieMax à partir des statistiques
// de base, des bonus de niveau et des bonus d'équipement. Toujours
// appelée après un changement de niveau ou d'équipement ; les PV/énergie
// courants ne sont jamais augmentés au passage, seulement plafonnés si
// le nouveau maximum est plus bas (objet retiré).
func (c *Character) RecomputeMaxStats() {
	c.PVMax = c.BasePVMax + c.NiveauBonusPV + c.Equip.TotalBonusPV()
	c.EnergieMax = c.BaseEnergieMax + c.NiveauBonusEN
	if c.PV > c.PVMax {
		c.PV = c.PVMax
	}
	if c.Energie > c.EnergieMax {
		c.Energie = c.EnergieMax
	}
}

// AddXP — mission bonus « expérience » : ajoute de l'XP et fait monter
// de niveau autant de fois que nécessaire (PV/énergie max augmentent,
// les PV/énergie courants ne sont pas rendus, seulement le plafond).
// Renvoie le nombre de niveaux gagnés.
func (c *Character) AddXP(xp int) int {
	c.XP += xp
	gained := 0
	for c.XP >= c.XPMax {
		c.XP -= c.XPMax
		c.Niveau++
		c.NiveauBonusPV += NiveauBonusPVParPalier
		c.NiveauBonusEN += NiveauBonusENParPalier
		c.XPMax += XPParNiveau
		gained++
	}
	if gained > 0 {
		c.RecomputeMaxStats()
	}
	return gained
}

// Die — tâche « mort puis résurrection à 50 % des PV ». Ne renvoie pas
// au menu : le personnage se réveille avec la moitié de ses PV max.
func (c *Character) Die() {
	c.PV = c.PVMax / 2
	c.Energie = c.EnergieMax
}

// EquipItem pose une pièce d'équipement, échange l'ancienne (renvoyée
// dans l'inventaire si possible) et recalcule les PV max.
func (c *Character) EquipItem(item *EquipmentItem) {
	old := c.Equip.Set(item)
	c.RecomputeMaxStats()
	if old != nil {
		c.AddItem(old.Nom)
	}
}

// FormatNom — tâche 11 : lettres uniquement, première lettre en majuscule,
// le reste en minuscules.
func FormatNom(s string) string {
	var b strings.Builder
	for _, r := range s {
		if unicode.IsLetter(r) {
			b.WriteRune(r)
		}
	}
	runes := []rune(strings.ToLower(b.String()))
	if len(runes) == 0 {
		return ""
	}
	runes[0] = unicode.ToUpper(runes[0])
	return string(runes)
}

// AddItem — tâche 12 : refuse l'objet si l'inventaire est plein.
func (c *Character) AddItem(item string) bool {
	if len(c.Inventaire) >= c.Capacite {
		return false
	}
	c.Inventaire = append(c.Inventaire, item)
	return true
}

// RemoveItem retire l'objet à l'emplacement i.
func (c *Character) RemoveItem(i int) {
	if i < 0 || i >= len(c.Inventaire) {
		return
	}
	c.Inventaire = append(c.Inventaire[:i], c.Inventaire[i+1:]...)
}

// UpgradeInventory — tâche 18 : +10 emplacements, 3 fois au maximum.
func (c *Character) UpgradeInventory() bool {
	if c.Ameliorations >= InventaireMaxUpgr {
		return false
	}
	c.Ameliorations++
	c.Capacite += InventaireBonus
	return true
}

// UseItem utilise l'objet de l'emplacement i et renvoie un message à afficher.
// C'est ici que se branche la logique du back (takePot, etc.).
func (c *Character) UseItem(i int) string {
	if i < 0 || i >= len(c.Inventaire) {
		return ""
	}
	name := c.Inventaire[i]
	switch name {
	case "Almond Water":
		if c.PV >= c.PVMax {
			return "Tes PV sont déjà au maximum."
		}
		avant := c.PV
		c.PV += SoinAlmondWater
		if c.PV > c.PVMax {
			c.PV = c.PVMax
		}
		c.RemoveItem(i)
		return "Almond Water bue : +" + strconv.Itoa(c.PV-avant) + " PV"
	case "Potion de poison":
		return "Se garde pour empoisonner un monstre en combat (Sac)."
	}
	// Pièce d'équipement rangée dans le sac (renvoyée par un échange
	// précédent chez le forgeron) : l'utiliser l'équipe à nouveau.
	if eq, ok := EquipmentCatalogue[name]; ok {
		c.RemoveItem(i)
		c.EquipItem(&eq)
		return name + " équipé(e) (" + eq.Slot.String() + ")."
	}
	return "Cet objet ne s'utilise pas directement."
}

// ItemDesc — description affichée dans l'inventaire.
var ItemDesc = map[string]string{
	"Almond Water":         "Eau légèrement sucrée. Rend 50 PV.",
	"Potion de poison":     "À jeter sur un monstre en combat : dégâts sur la durée.",
	"Morceau de moquette":  "Humide, jaunâtre. Matériau d'artisanat.",
	"Tuyau rouillé":        "Arraché à un mur du Level 2.",
	"Néon cassé":           "Il grésille encore un peu.",
	"Barre de fer":         "Lourde et froide. Mieux que rien.",
	"Casque en tuyau":      "Fabriqué par le forgeron. +20 PV max.",
	"Plastron de moquette": "Fabriqué par le forgeron. +35 PV max.",
	"Bottes en fer":        "Fabriquées par le forgeron. +15 PV max.",
}
