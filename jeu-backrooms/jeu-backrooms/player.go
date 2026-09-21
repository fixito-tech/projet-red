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
}

// ClassInfo décrit une classe jouable.
// Base = la classe de l'énoncé dont elle reprend les statistiques.
type ClassInfo struct {
	Nom        string
	Base       string
	PVMax      int
	EnergieMax int
	Desc       string
}

// Classes — tâche 11 : Humain 100 PV max, Elfe 80, Nain 120.
// L'énergie n'est pas fixée par l'énoncé : ce sont nos choix d'équilibrage.
var Classes = []ClassInfo{
	{"Survivant", "Humain", 100, 100,
		"Équilibré. Il a tenu jusqu'ici, il tiendra encore un peu."},
	{"Ancien Résident", "Elfe", 80, 120,
		"Fragile, mais il connaît les failles entre les niveaux."},
	{"Chasseur de niveaux", "Nain", 120, 80,
		"Encaisse les coups des entités. Se fatigue vite."},
}

const (
	InventaireBase    = 10 // tâche 12
	InventaireBonus   = 10 // tâche 18
	InventaireMaxUpgr = 3  // tâche 18
	SoinAlmondWater   = 50 // tâche 5 : la potion rend 50 PV
)

// NewCharacter — tâche 11 : niveau 1, PV actuels = 50 % des PV max.
func NewCharacter(nom string, c ClassInfo) *Character {
	return &Character{
		Nom:        nom,
		Classe:     c.Nom,
		Niveau:     1,
		PVMax:      c.PVMax,
		PV:         c.PVMax / 2,
		EnergieMax: c.EnergieMax,
		Energie:    c.EnergieMax,
		Inventaire: []string{"Almond Water", "Almond Water", "Almond Water"},
		Capacite:   InventaireBase,
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
	switch c.Inventaire[i] {
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
	}
	return "Cet objet ne s'utilise pas directement."
}

// ItemDesc — description affichée dans l'inventaire.
var ItemDesc = map[string]string{
	"Almond Water":        "Eau légèrement sucrée. Rend 50 PV.",
	"Morceau de moquette": "Humide, jaunâtre. Matériau d'artisanat.",
	"Tuyau rouillé":       "Arraché à un mur du Level 2.",
	"Néon cassé":          "Il grésille encore un peu.",
	"Barre de fer":        "Lourde et froide. Mieux que rien.",
}
