package main

// ---------------------------------------------------------------
// CLAVIER DU MENU TITRE ET DE LA CRÉATION DU PERSONNAGE (tâche 11),
// plus l'écran de victoire. Le dessin de ces écrans est dans
// menu_ui.go.
// ---------------------------------------------------------------

import (
	"unicode"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

// updateMenu gère le menu titre : choisir avec les flèches, valider avec Entrée.
func (g *Game) updateMenu() error {
	items := g.menuEntries()
	if g.menuIndex >= len(items) {
		g.menuIndex = 0
	}
	navigateList(&g.menuIndex, len(items))
	if validatePressed() {
		switch items[g.menuIndex] {
		case "Continuer":
			g.state = StatePlay
		case "Jouer", "Nouvelle partie":
			g.state = StateCreate
			g.createStep = 0
			g.nameBuf = nil
			g.classSel = 0
		case "Quitter":
			return ebiten.Termination
		}
	}
	return nil
}

// menuEntries : "Continuer" n'apparaît qu'une fois un personnage créé.
func (g *Game) menuEntries() []string {
	if g.hero != nil {
		return []string{"Continuer", "Nouvelle partie", "Quitter"}
	}
	return []string{"Jouer", "Quitter"}
}

// updateCreate : écran de création du personnage (tâche 11), en deux
// étapes : d'abord le nom, puis la classe.
func (g *Game) updateCreate() error {
	if g.createStep == 0 {
		g.updateNameEntry()
	} else {
		g.updateClassChoice()
	}
	return nil
}

// updateNameEntry : saisie du nom, en ne gardant que les lettres.
func (g *Game) updateNameEntry() {
	for _, r := range ebiten.AppendInputChars(nil) {
		if unicode.IsLetter(r) && len(g.nameBuf) < maxNameLength {
			g.nameBuf = append(g.nameBuf, r)
		}
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyBackspace) && len(g.nameBuf) > 0 {
		g.nameBuf = g.nameBuf[:len(g.nameBuf)-1]
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		g.state = StateMenu
	}
	if confirmPressed() && FormatNom(string(g.nameBuf)) != "" {
		g.createStep = 1
	}
}

// updateClassChoice : choix de la classe avec gauche/droite, puis la
// validation crée le personnage et lance la partie.
func (g *Game) updateClassChoice() {
	if leftPressed() {
		g.classSel = (g.classSel - 1 + len(Classes)) % len(Classes)
	}
	if rightPressed() {
		g.classSel = (g.classSel + 1) % len(Classes)
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		g.createStep = 0
	}
	if validatePressed() {
		g.hero = NewCharacter(FormatNom(string(g.nameBuf)), Classes[g.classSel])
		g.startNewGame()
		g.state = StatePlay
	}
}

// updateVictory gère l'écran de victoire.
func (g *Game) updateVictory() error {
	if validatePressed() {
		g.state = StateMenu
	}
	return nil
}
