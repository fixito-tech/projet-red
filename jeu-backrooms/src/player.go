package main

// ---------------------------------------------------------------
// PERSONNAGE — données uniquement, aucun affichage ici.
//
// Ce fichier suit la struct Character de l'énoncé. Si la struct du back
// (tes collègues) porte d'autres noms de champs, c'est UNIQUEMENT ici qu'il
// faut adapter : l'interface (fichiers *_ui.go) ne lit que ces champs-là.
// ---------------------------------------------------------------

import (
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
	c.PV = min(c.PV, c.PVMax)
	c.Energie = min(c.Energie, c.EnergieMax)
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

// heal rend des PV sans dépasser le maximum et renvoie le nombre de PV
// réellement rendus (utilisé par l'Almond Water et le sort de soin).
func (c *Character) heal(amount int) int {
	before := c.PV
	c.PV = min(c.PV+amount, c.PVMax)
	return c.PV - before
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
