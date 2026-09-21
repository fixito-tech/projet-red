package main

import "fmt"

func main() {
	useColor = enableANSI()
	printBanner()
	printTitle("BIENVENUE DANS LES BACKROOMS")
	fmt.Println("Vous venez de \"no-clip\" hors de la réalité.")
	fmt.Println("Survivez, explorez, et retrouvez votre chemin...")
	pause()

	character := characterCreation()
	mainMenu(&character)

	fmt.Println("\nMerci d'avoir joué. À bientôt dans les Backrooms...")
}

// mainMenu affiche le menu principal du jeu et redirige vers les différentes fonctionnalités.
func mainMenu(c *Character) {
	for {
		printTitle("MENU PRINCIPAL")
		fmt.Println("1. Informations du personnage")
		fmt.Println("2. Inventaire")
		fmt.Println("3. Le Troqueur")
		fmt.Println("4. Le Bricoleur")
		fmt.Println("5. Entraînement")
		fmt.Println("6. Qui sont-ils ?")
		fmt.Println("7. Quêtes")
		fmt.Println("8. Explorer la carte")
		fmt.Println("0. Quitter")

		choice := readChoice("\nVotre choix : ")
		switch choice {
		case 1:
			displayInfo(c)
			pause()
		case 2:
			accessInventory(c)
		case 3:
			accessTroqueur(c)
		case 4:
			accessBricoleur(c)
		case 5:
			trainingFight(c)
		case 6:
			whoAreThey()
			pause()
		case 7:
			accessQuests(c)
			pause()
		case 8:
			exploreMap(c)
		case 0:
			return
		default:
			fmt.Println("Choix invalide.")
		}
	}
}

// whoAreThey révèle les artistes cachés dans les intitulés des parties 2 et 3 du sujet.
func whoAreThey() {
	printTitle("QUI SONT-ILS ?")
	fmt.Println("Partie 02 (Économie & fabrication) : les titres des tâches citent des")
	fmt.Println("chansons du groupe ABBA -> Money, Money, Money ; Gimme! Gimme! Gimme! ;")
	fmt.Println("Mamma Mia ; On and On and On ; Two for the Price of One.")
	fmt.Println()
	fmt.Println("Partie 03 (Combat tour par tour) : les titres des tâches citent des")
	fmt.Println("films de Steven Spielberg -> Duel ; A.I. Intelligence Artificielle ;")
	fmt.Println("Ready Player One.")
}
