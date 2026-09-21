package main

import (
	"fmt"
	"strings"
	"unicode"
)

const colorGray = "\033[90m"

// worldMap dessine les couloirs explorables :
// # mur, . sol, T le Troqueur, B le Bricoleur, H un Hurleur (combat d'entraînement).
var worldMap = []string{
	"####################",
	"#..................#",
	"#.T....#####....B..#",
	"#......#...#.......#",
	"#..H...#...#....H..#",
	"#......##.##.......#",
	"#..................#",
	"#........H.........#",
	"####################",
}

// Position de départ du personnage sur la carte.
const (
	startX = 1
	startY = 1
)

// tileAt retourne le caractère de la case (x, y) ; hors de la carte, c'est un mur.
func tileAt(x, y int) byte {
	if y < 0 || y >= len(worldMap) || x < 0 || x >= len(worldMap[y]) {
		return '#'
	}
	return worldMap[y][x]
}

// moveOffset convertit une touche de déplacement en décalage (dx, dy).
func moveOffset(key rune) (dx, dy int, ok bool) {
	switch key {
	case 'z', 'w':
		return 0, -1, true
	case 's':
		return 0, 1, true
	case 'q', 'a':
		return -1, 0, true
	case 'd':
		return 1, 0, true
	}
	return 0, 0, false
}

// tileSymbol retourne le symbole coloré d'une case de la carte.
func tileSymbol(tile byte) string {
	switch tile {
	case '#':
		return paint(colorGray, "█")
	case 'T':
		return paint(colorYellow, "T")
	case 'B':
		return paint(colorCyan, "B")
	case 'H':
		return paint(colorRed, "H")
	}
	return paint(colorGray, "·")
}

// drawMap affiche la carte avec le personnage à la position (px, py).
func drawMap(px, py int) {
	if useColor {
		fmt.Print("\033[2J\033[H")
	}
	printTitle("CARTE : LES COULOIRS JAUNES")
	for y, row := range worldMap {
		var line strings.Builder
		for x := 0; x < len(row); x++ {
			if x == px && y == py {
				line.WriteString(paint(colorBold+colorGreen, "@"))
				continue
			}
			line.WriteString(tileSymbol(row[x]))
		}
		fmt.Println(line.String())
	}
	fmt.Printf("\n%s vous   %s Troqueur   %s Bricoleur   %s Hurleur (combat)\n",
		paint(colorGreen, "@"), tileSymbol('T'), tileSymbol('B'), tileSymbol('H'))
}

// isPlace indique si la case déclenche une action (boutique, forge, combat).
func isPlace(tile byte) bool {
	return tile == 'T' || tile == 'B' || tile == 'H'
}

// interact déclenche l'action de la case sur laquelle se trouve le personnage.
func interact(c *Character, tile byte) {
	switch tile {
	case 'T':
		accessTroqueur(c)
	case 'B':
		accessBricoleur(c)
	case 'H':
		trainingFight(c)
	}
}

// exploreMap laisse le joueur déplacer son personnage sur la carte.
// Sous Windows, les touches sont lues en direct (sans Entrée) ; sinon on saisit
// les déplacements puis Entrée.
func exploreMap(c *Character) {
	x, y := startX, startY
	live := setRawInput(true)
	if live {
		defer setRawInput(false)
	}

	message := ""
	for {
		drawMap(x, y)
		if message != "" {
			fmt.Println(message)
			message = ""
		}

		var keys []rune
		if live {
			fmt.Println("Déplacements : z/w haut, s bas, q/a gauche, d droite. X pour quitter.")
			b, err := reader.ReadByte()
			if err != nil {
				return
			}
			key := unicode.ToLower(rune(b))
			if key == 'x' {
				return
			}
			keys = []rune{key}
		} else {
			fmt.Println("Déplacements : z/w haut, s bas, q/a gauche, d droite (ex : zzdd). 0 pour quitter.")
			input := readLine("> ")
			if input == "0" {
				return
			}
			keys = []rune(strings.ToLower(input))
		}

		for _, key := range keys {
			dx, dy, ok := moveOffset(key)
			if !ok {
				continue
			}
			if tileAt(x+dx, y+dy) == '#' {
				message = "Un mur vous bloque le passage."
				break
			}
			x, y = x+dx, y+dy
			if isPlace(tileAt(x, y)) {
				if live {
					setRawInput(false)
				}
				interact(c, tileAt(x, y))
				if live {
					setRawInput(true)
				}
				break
			}
		}
	}
}
