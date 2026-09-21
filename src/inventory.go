package main

import (
	"fmt"
	"time"
)

// addInventory ajoute un objet à l'inventaire du personnage si une place est disponible.
func addInventory(c *Character, itemName string) bool {
	if len(c.Inventory) >= c.MaxSlots {
		fmt.Printf("Inventaire plein (%d/%d) ! Impossible d'ajouter %s.\n", len(c.Inventory), c.MaxSlots, itemName)
		return false
	}
	c.Inventory = append(c.Inventory, itemName)
	return true
}

// removeInventory retire une occurrence de l'objet de l'inventaire du personnage.
func removeInventory(c *Character, itemName string) bool {
	for i, item := range c.Inventory {
		if item == itemName {
			c.Inventory = append(c.Inventory[:i], c.Inventory[i+1:]...)
			return true
		}
	}
	fmt.Printf("%s n'est pas dans l'inventaire.\n", itemName)
	return false
}

// countItem compte le nombre d'occurrences d'un objet dans l'inventaire.
func countItem(c *Character, itemName string) int {
	count := 0
	for _, item := range c.Inventory {
		if item == itemName {
			count++
		}
	}
	return count
}

// accessInventory affiche le contenu de l'inventaire et permet d'utiliser un objet.
func accessInventory(c *Character) {
	for {
		printTitle("INVENTAIRE")
		if len(c.Inventory) == 0 {
			fmt.Println("Votre inventaire est vide.")
		}

		unique := uniqueItems(c.Inventory)
		for i, item := range unique {
			fmt.Printf("%d. %s x%d\n", i+1, item, countItem(c, item))
		}
		fmt.Println("0. Retour")

		choice := readChoice("\nUtiliser quel objet ? ")
		if choice == 0 {
			return
		}
		if choice < 1 || choice > len(unique) {
			fmt.Println("Choix invalide.")
			pause()
			continue
		}

		useItem(c, unique[choice-1])
		pause()
	}
}

// uniqueItems retourne la liste des objets distincts présents dans l'inventaire.
func uniqueItems(inventory []string) []string {
	seen := map[string]bool{}
	var unique []string
	for _, item := range inventory {
		if !seen[item] {
			seen[item] = true
			unique = append(unique, item)
		}
	}
	return unique
}

// useItem applique l'effet d'un objet de l'inventaire lorsqu'il est utilisé.
func useItem(c *Character, itemName string) {
	switch itemName {
	case "Eau d'Amande":
		takePot(c)
	case "Eau Croupie":
		poisonPot(c)
	case "Stimulant":
		useStimulant(c)
	case "Manuel : Cocktail Molotov":
		spellBook(c)
	case "Sac à Dos Amélioré":
		upgradeInventorySlot(c)
	case "Casque de Chantier", "Veste Matelassée", "Bottes de Randonnée":
		equip(c, itemName)
	default:
		fmt.Printf("%s ne peut pas être utilisé directement (matériau de fabrication).\n", itemName)
	}
}

// takePot consomme une Eau d'Amande : régénère 50 PV sans dépasser le maximum.
func takePot(c *Character) {
	if !removeInventory(c, "Eau d'Amande") {
		return
	}
	c.CurrentHP += 50
	if c.CurrentHP > c.MaxHP {
		c.CurrentHP = c.MaxHP
	}
	fmt.Printf("%s boit une Eau d'Amande.\n", c.Name)
	fmt.Printf("PV : %d / %d\n", c.CurrentHP, c.MaxHP)
}

// poisonPot consomme une Eau Croupie : inflige 10 dégâts par seconde pendant 3 secondes.
func poisonPot(c *Character) {
	if !removeInventory(c, "Eau Croupie") {
		return
	}
	fmt.Printf("%s boit une Eau Croupie sans le savoir...\n", c.Name)
	for i := 0; i < 3; i++ {
		time.Sleep(1 * time.Second)
		c.CurrentHP -= 10
		if c.CurrentHP < 0 {
			c.CurrentHP = 0
		}
		fmt.Printf("PV : %d / %d\n", c.CurrentHP, c.MaxHP)
		if c.CurrentHP == 0 {
			break
		}
	}
}

// useStimulant consomme un Stimulant : régénère 15 points de Mana (Adrénaline).
func useStimulant(c *Character) {
	if !removeInventory(c, "Stimulant") {
		return
	}
	c.Mana += 15
	if c.Mana > c.MaxMana {
		c.Mana = c.MaxMana
	}
	fmt.Printf("%s utilise un Stimulant.\n", c.Name)
	fmt.Printf("Mana : %d / %d\n", c.Mana, c.MaxMana)
}

// spellBook apprend le sort "Cocktail Molotov" au personnage (une seule fois).
func spellBook(c *Character) {
	for _, skill := range c.Skills {
		if skill == "Cocktail Molotov" {
			fmt.Println("Vous connaissez déjà ce sort, le manuel reste dans votre inventaire.")
			return
		}
	}
	if !removeInventory(c, "Manuel : Cocktail Molotov") {
		return
	}
	c.Skills = append(c.Skills, "Cocktail Molotov")
	fmt.Printf("%s apprend le sort Cocktail Molotov !\n", c.Name)
}

// upgradeInventorySlot consomme un Sac à Dos Amélioré et augmente la capacité de
// l'inventaire de +10, utilisable 3 fois maximum.
func upgradeInventorySlot(c *Character) {
	if c.SlotUpgrades >= 3 {
		fmt.Println("Votre sac à dos ne peut plus être amélioré (limite de 3 atteinte).")
		return
	}
	if !removeInventory(c, "Sac à Dos Amélioré") {
		return
	}
	c.MaxSlots += 10
	c.SlotUpgrades++
	fmt.Printf("Capacité de l'inventaire augmentée : %d emplacements (%d/3 améliorations utilisées).\n", c.MaxSlots, c.SlotUpgrades)
}
