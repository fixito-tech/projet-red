package main

import "fmt"

// characterTurn simule le tour de jeu du personnage face à un monstre.
func characterTurn(c *Character, m *Monster) {
	for {
		printTitle("TOUR DE " + c.Name)
		fmt.Println("1. Attaquer")
		fmt.Println("2. Sorts")
		fmt.Println("3. Inventaire")

		choice := readChoice("\nQue faites-vous ? ")
		switch choice {
		case 1:
			basicAttack(c, m)
			return
		case 2:
			if castSpell(c, m) {
				return
			}
		case 3:
			if combatInventory(c) {
				return
			}
		default:
			fmt.Println("Choix invalide.")
		}
	}
}

// basicAttack applique l'Attaque basique du personnage sur le monstre ciblé.
func basicAttack(c *Character, m *Monster) {
	damage := 5
	m.CurrentHP -= damage
	if m.CurrentHP < 0 {
		m.CurrentHP = 0
	}
	fmt.Printf("\n%s utilise Attaque basique.\n", c.Name)
	fmt.Printf("%s inflige %d dégâts à %s.\n", c.Name, damage, m.Name)
	fmt.Printf("%s : PV %d / %d\n", m.Name, m.CurrentHP, m.MaxHP)
}

// castSpell affiche les sorts connus et lance celui choisi sur le monstre.
// Retourne true si un sort a été lancé (le tour est alors terminé).
func castSpell(c *Character, m *Monster) bool {
	printTitle("SORTS")
	for i, skill := range c.Skills {
		fmt.Printf("%d. %s (%d dégâts, %d mana)\n", i+1, skill, spellDamage[skill], spellCost[skill])
	}
	fmt.Println("0. Retour")

	choice := readChoice("\nQuel sort lancer ? ")
	if choice == 0 || choice < 1 || choice > len(c.Skills) {
		return false
	}

	skill := c.Skills[choice-1]
	cost := spellCost[skill]
	if c.Mana < cost {
		fmt.Println("Mana insuffisant pour lancer ce sort.")
		pause()
		return false
	}

	c.Mana -= cost
	damage := spellDamage[skill]
	m.CurrentHP -= damage
	if m.CurrentHP < 0 {
		m.CurrentHP = 0
	}
	fmt.Printf("\n%s lance %s.\n", c.Name, skill)
	fmt.Printf("%s inflige %d dégâts à %s.\n", c.Name, damage, m.Name)
	fmt.Printf("%s : PV %d / %d\n", m.Name, m.CurrentHP, m.MaxHP)
	return true
}

// combatInventory affiche l'inventaire et permet d'utiliser un objet pendant le combat.
// Retourne true si un objet a été utilisé (le tour est alors terminé).
func combatInventory(c *Character) bool {
	printTitle("INVENTAIRE (COMBAT)")
	if len(c.Inventory) == 0 {
		fmt.Println("Votre inventaire est vide.")
		pause()
		return false
	}

	unique := uniqueItems(c.Inventory)
	for i, item := range unique {
		fmt.Printf("%d. %s x%d\n", i+1, item, countItem(c, item))
	}
	fmt.Println("0. Retour")

	choice := readChoice("\nUtiliser quel objet ? ")
	if choice == 0 || choice < 1 || choice > len(unique) {
		return false
	}

	useItem(c, unique[choice-1])
	return true
}

// trainingFight lance un combat d'entraînement tour par tour entre le personnage et un Hurleur d'entraînement.
func trainingFight(c *Character) {
	monster := initGoblin()
	turn := 1

	printTitle("ENTRAÎNEMENT")
	fmt.Printf("%s affronte %s (PV : %d, Attaque : %d) !\n", c.Name, monster.Name, monster.MaxHP, monster.Attack)

	playerFirst := c.Initiative >= monster.Initiative
	if playerFirst {
		fmt.Printf("%s a l'initiative et commence le combat.\n", c.Name)
	} else {
		fmt.Printf("%s a l'initiative et commence le combat.\n", monster.Name)
	}
	pause()

	for {
		fmt.Printf("\n=== Tour %d ===\n", turn)

		if playerFirst {
			characterTurn(c, &monster)
			if monster.CurrentHP <= 0 {
				endTrainingFight(c, &monster, true)
				return
			}
			goblinPattern(&monster, c, turn)
			if isDead(c) {
				endTrainingFight(c, &monster, false)
				return
			}
		} else {
			goblinPattern(&monster, c, turn)
			if isDead(c) {
				endTrainingFight(c, &monster, false)
				return
			}
			characterTurn(c, &monster)
			if monster.CurrentHP <= 0 {
				endTrainingFight(c, &monster, true)
				return
			}
		}

		pause()
		turn++
	}
}

// endTrainingFight affiche le résultat du combat et récompense le joueur en cas de victoire.
func endTrainingFight(c *Character, m *Monster, victory bool) {
	if victory {
		fmt.Printf("\n%s est vaincu !\n", m.Name)
		gainExperience(c, m.ExpReward)
		updateQuests(c, "kill", 1)
	} else {
		fmt.Println("\nLe combat est terminé.")
	}
	fmt.Println("Retour au menu principal.")
	pause()
}
