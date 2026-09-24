package main

// ---------------------------------------------------------------
// INVENTAIRE — règles du sac du personnage : ajouter, retirer,
// agrandir, utiliser un objet, compter/chercher un objet. Logique
// pure : le dessin est dans inventory_ui.go.
// ---------------------------------------------------------------

import "strconv"

const (
	InventaireBase    = 10 // tâche 12
	InventaireBonus   = 10 // tâche 18
	InventaireMaxUpgr = 3  // tâche 18
	SoinAlmondWater   = 50 // tâche 5 : la potion rend 50 PV
)

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
		soin := c.heal(SoinAlmondWater)
		c.RemoveItem(i)
		return "Almond Water bue : +" + strconv.Itoa(soin) + " PV"
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

// removeItems retire qty exemplaires de l'objet name de l'inventaire,
// un par un. S'il n'y en a plus assez, s'arrête sans rien casser (le
// forgeron vérifie les quantités avant d'appeler cette fonction).
func (c *Character) removeItems(name string, qty int) {
	for n := 0; n < qty; n++ {
		i := firstIndexOf(c.Inventaire, name)
		if i < 0 {
			return
		}
		c.RemoveItem(i)
	}
}

// countItem compte le nombre d'exemplaires d'un objet dans l'inventaire.
func countItem(inv []string, name string) int {
	n := 0
	for _, it := range inv {
		if it == name {
			n++
		}
	}
	return n
}

// firstIndexOf renvoie l'index du premier objet nommé "name" dans
// l'inventaire (utilisé pour traduire une sélection du sac de combat
// en emplacement réel de l'inventaire).
func firstIndexOf(inv []string, name string) int {
	for i, it := range inv {
		if it == name {
			return i
		}
	}
	return -1
}
