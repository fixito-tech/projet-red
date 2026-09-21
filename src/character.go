package main

import (
	"fmt"
	"math/rand"
	"regexp"
)

var lettersOnly = regexp.MustCompile(`^[A-Za-zÀ-ÖØ-öø-ÿ]+$`)

// initCharacter initialise un personnage avec les valeurs fournies.
func initCharacter(name, class string, level, maxHP, currentHP int, inventory []string) Character {
	return Character{
		Name:          name,
		Class:         class,
		Level:         level,
		MaxHP:         maxHP,
		CurrentHP:     currentHP,
		Inventory:     inventory,
		MaxSlots:      10,
		Skills:        []string{"Coup de Poing"},
		Gold:          100,
		Equipment:     Equipment{},
		Initiative:    5 + rand.Intn(10),
		Experience:    0,
		ExperienceMax: 100,
		Mana:          20,
		MaxMana:       20,
	}
}

// characterCreation laisse l'utilisateur choisir le nom et la classe de son survivant.
func characterCreation() Character {
	printTitle("CRÉATION DU SURVIVANT")

	var name string
	for {
		input := readLine("Entrez le nom de votre survivant (lettres uniquement) : ")
		if input != "" && lettersOnly.MatchString(input) {
			name = title(input)
			break
		}
		fmt.Println("Nom invalide : uniquement des lettres, merci de réessayer.")
	}

	fmt.Println("\nChoisissez votre classe :")
	fmt.Println("1. Équilibré (100 PV max)")
	fmt.Println("2. Éclaireur (80 PV max, plus rapide)")
	fmt.Println("3. Robuste (120 PV max, plus résistant)")

	var class string
	var maxHP int
	for {
		choice := readChoice("Votre choix : ")
		switch choice {
		case 1:
			class, maxHP = "Équilibré", 100
		case 2:
			class, maxHP = "Éclaireur", 80
		case 3:
			class, maxHP = "Robuste", 120
		default:
			fmt.Println("Choix invalide, réessayez.")
			continue
		}
		break
	}

	c := initCharacter(name, class, 1, maxHP, maxHP/2, []string{})
	fmt.Printf("\nBienvenue dans les Backrooms, %s !\n", c.Name)
	fmt.Printf("Classe : %s\n", c.Class)
	pause()
	return c
}

// displayInfo affiche les informations complètes du personnage.
func displayInfo(c *Character) {
	printTitle("FICHE DU SURVIVANT")
	fmt.Printf("Nom          : %s\n", c.Name)
	fmt.Printf("Classe       : %s\n", c.Class)
	fmt.Printf("Niveau       : %d\n", c.Level)
	fmt.Printf("PV           : %d / %d\n", c.CurrentHP, c.MaxHP)
	fmt.Printf("Mana         : %d / %d\n", c.Mana, c.MaxMana)
	fmt.Printf("Expérience   : %d / %d\n", c.Experience, c.ExperienceMax)
	fmt.Printf("Initiative   : %d\n", c.Initiative)
	fmt.Printf("Jetons       : %d\n", c.Gold)
	fmt.Printf("Sorts appris : %v\n", c.Skills)
	fmt.Println("Équipement :")
	fmt.Printf("  Tête  : %s\n", emptyIfBlank(c.Equipment.Tete))
	fmt.Printf("  Torse : %s\n", emptyIfBlank(c.Equipment.Torse))
	fmt.Printf("  Pieds : %s\n", emptyIfBlank(c.Equipment.Pieds))
	fmt.Printf("Inventaire (%d/%d) : %v\n", len(c.Inventory), c.MaxSlots, c.Inventory)
	printLine()
}

func emptyIfBlank(s string) string {
	if s == "" {
		return "(aucun)"
	}
	return s
}

// isDead vérifie si le personnage est tombé à 0 PV ou moins.
// Si c'est le cas, il s'évanouit puis reprend connaissance avec 50% de ses PV max.
func isDead(c *Character) bool {
	if c.CurrentHP > 0 {
		return false
	}
	c.CurrentHP = c.MaxHP / 2
	fmt.Printf("\n%s s'effondre, à bout de forces...\n", c.Name)
	fmt.Println("Après un vide noir, il/elle reprend connaissance dans un autre niveau.")
	fmt.Printf("%s se réveille avec %d / %d PV.\n", c.Name, c.CurrentHP, c.MaxHP)
	return true
}

// gainExperience ajoute de l'expérience au personnage et gère la montée de niveau.
func gainExperience(c *Character, amount int) {
	fmt.Printf("\n%s gagne %d points d'expérience.\n", c.Name, amount)
	c.Experience += amount
	for c.Experience >= c.ExperienceMax {
		c.Experience -= c.ExperienceMax
		c.Level++
		bonusHP := 10
		c.MaxHP += bonusHP
		c.CurrentHP = c.MaxHP
		c.ExperienceMax = int(float64(c.ExperienceMax) * 1.5)
		fmt.Printf("%s passe niveau %d ! (+%d PV max, PV restaurés)\n", c.Name, c.Level, bonusHP)
	}
}

// equip équipe un équipement fabriqué dans le bon emplacement du personnage.
func equip(c *Character, itemName string) {
	info, ok := equipmentCatalog[itemName]
	if !ok {
		fmt.Println("Cet objet ne peut pas être équipé.")
		return
	}

	if !removeInventory(c, itemName) {
		return
	}

	var current *string
	switch info.Slot {
	case "Tete":
		current = &c.Equipment.Tete
	case "Torse":
		current = &c.Equipment.Torse
	case "Pieds":
		current = &c.Equipment.Pieds
	}

	if *current != "" {
		oldInfo := equipmentCatalog[*current]
		c.MaxHP -= oldInfo.HPBonus
		if !addInventory(c, *current) {
			// Sécurité : ne devrait pas arriver car on vient de libérer une place.
			fmt.Println("Impossible de récupérer l'ancien équipement, inventaire plein.")
		} else {
			fmt.Printf("%s récupère %s dans son inventaire.\n", c.Name, *current)
		}
	}

	*current = itemName
	c.MaxHP += info.HPBonus
	if c.CurrentHP > c.MaxHP {
		c.CurrentHP = c.MaxHP
	}
	fmt.Printf("%s équipe %s (+%d PV max).\n", c.Name, itemName, info.HPBonus)
}
