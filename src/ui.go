package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

var reader = bufio.NewReader(os.Stdin)

// readLine lit une ligne saisie par l'utilisateur et retire les espaces superflus.
// Si l'entrée standard est fermée (EOF), le programme s'arrête proprement.
func readLine(prompt string) string {
	fmt.Print(prompt)
	line, err := reader.ReadString('\n')
	if err != nil && line == "" {
		fmt.Println("\nFin de la saisie, fermeture du jeu.")
		os.Exit(0)
	}
	return strings.TrimSpace(line)
}

// readChoice lit un choix numérique saisi par l'utilisateur dans un menu.
func readChoice(prompt string) int {
	input := readLine(prompt)
	choice, err := strconv.Atoi(input)
	if err != nil {
		return -1
	}
	return choice
}

// pause attend que l'utilisateur appuie sur Entrée avant de continuer.
func pause() {
	readLine("\nAppuyez sur Entrée pour continuer...")
}

// title formate une chaîne avec une majuscule initiale et le reste en minuscule.
func title(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

// printLine affiche une ligne de séparation pour aérer l'affichage.
func printLine() {
	fmt.Println(strings.Repeat("-", 46))
}

// printTitle affiche un titre encadré.
func printTitle(t string) {
	printLine()
	fmt.Println(t)
	printLine()
}
