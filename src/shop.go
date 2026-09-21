package main

import "fmt"

// accessTroqueur affiche le menu du Troqueur (le marchand des Backrooms) et gère les achats.
func accessTroqueur(c *Character) {
	for {
		printTitle("LE TROQUEUR")
		fmt.Printf("Jetons disponibles : %d\n\n", c.Gold)
		for i, name := range shopOrder {
			info := itemCatalog[name]
			fmt.Printf("%d. %s - %d jeton(s)\n", i+1, info.Name, info.Price)
		}
		fmt.Println("0. Retour")

		choice := readChoice("\nQue voulez-vous échanger ? ")
		if choice == 0 {
			return
		}
		if choice < 1 || choice > len(shopOrder) {
			fmt.Println("Choix invalide.")
			pause()
			continue
		}

		buyItem(c, shopOrder[choice-1])
		pause()
	}
}

// buyItem tente d'acheter un objet chez le Troqueur.
func buyItem(c *Character, itemName string) {
	info := itemCatalog[itemName]
	if c.Gold < info.Price {
		fmt.Printf("Vous n'avez pas assez de jetons pour %s (%d requis, %d disponibles).\n", itemName, info.Price, c.Gold)
		return
	}
	if !addInventory(c, itemName) {
		return
	}
	c.Gold -= info.Price
	fmt.Printf("Vous obtenez %s.\n", itemName)
}

// accessBricoleur affiche le menu du Bricoleur (le forgeron des Backrooms) et gère la fabrication.
func accessBricoleur(c *Character) {
	for {
		printTitle("LE BRICOLEUR")
		fmt.Printf("Jetons disponibles : %d\n\n", c.Gold)
		for i, name := range equipmentOrder {
			info := equipmentCatalog[name]
			fmt.Printf("%d. %s (+%d PV max) - %d jeton(s)\n", i+1, info.Name, info.HPBonus, info.Cost)
			for material, qty := range info.Recipe {
				fmt.Printf("     - %d x %s (possédé(s) : %d)\n", qty, material, countItem(c, material))
			}
		}
		fmt.Println("0. Retour")

		choice := readChoice("\nQue voulez-vous fabriquer ? ")
		if choice == 0 {
			return
		}
		if choice < 1 || choice > len(equipmentOrder) {
			fmt.Println("Choix invalide.")
			pause()
			continue
		}

		craftEquipment(c, equipmentOrder[choice-1])
		pause()
	}
}

// craftEquipment fabrique un équipement si le personnage a les ressources, les jetons et la place nécessaires.
func craftEquipment(c *Character, itemName string) {
	info := equipmentCatalog[itemName]

	for material, qty := range info.Recipe {
		if countItem(c, material) < qty {
			fmt.Printf("Ressources insuffisantes : il vous faut %d x %s.\n", qty, material)
			return
		}
	}
	if c.Gold < info.Cost {
		fmt.Printf("Vous n'avez pas assez de jetons pour fabriquer %s (%d requis).\n", itemName, info.Cost)
		return
	}
	if len(c.Inventory) >= c.MaxSlots {
		fmt.Println("Inventaire plein : impossible de récupérer l'équipement fabriqué.")
		return
	}

	for material, qty := range info.Recipe {
		for i := 0; i < qty; i++ {
			removeInventory(c, material)
		}
	}
	c.Gold -= info.Cost
	addInventory(c, itemName)
	fmt.Printf("%s fabrique %s !\n", c.Name, itemName)
}
