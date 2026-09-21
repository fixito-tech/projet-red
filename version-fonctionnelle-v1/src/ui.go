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

// Codes couleur ANSI utilisés pour l'affichage.
const (
	colorReset  = "\033[0m"
	colorBold   = "\033[1m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorCyan   = "\033[36m"
)

// useColor active les couleurs ; il est mis à true au lancement si la console les gère.
var useColor = false

// paint colorie une chaîne si les couleurs sont activées.
func paint(color, s string) string {
	if !useColor {
		return s
	}
	return color + s + colorReset
}

// hpBar retourne une barre de vie colorée suivie de "actuels / max".
func hpBar(current, max, width int) string {
	if max <= 0 {
		max = 1
	}
	filled := current * width / max
	if filled < 0 {
		filled = 0
	}
	if filled > width {
		filled = width
	}

	color := colorGreen
	switch percent := current * 100 / max; {
	case percent <= 25:
		color = colorRed
	case percent <= 50:
		color = colorYellow
	}

	bar := strings.Repeat("█", filled) + strings.Repeat("░", width-filled)
	return fmt.Sprintf("[%s] %d / %d", paint(color, bar), current, max)
}

// printLine affiche une ligne de séparation pour aérer l'affichage.
func printLine() {
	fmt.Println(paint(colorCyan, strings.Repeat("═", 46)))
}

// printTitle affiche un titre encadré.
func printTitle(t string) {
	printLine()
	fmt.Println(paint(colorBold+colorYellow, t))
	printLine()
}

// printBanner affiche la bannière d'accueil du jeu.
func printBanner() {
	banner := []string{
		` ___   _   ___ _  _____  ___   ___   __  __ ___ `,
		`| _ ) /_\ / __| |/ / _ \/ _ \ / _ \ |  \/  / __|`,
		`| _ \/ _ \ (__| ' <|   / (_) | (_) || |\/| \__ \`,
		`|___/_/ \_\___|_|\_\_|_\\___/ \___/ |_|  |_|___/`,
	}
	for _, line := range banner {
		fmt.Println(paint(colorYellow, line))
	}
}
