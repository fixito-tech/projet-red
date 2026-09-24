package main

// ---------------------------------------------------------------
// TOUCHES DES MENUS — petites fonctions qui disent si une touche
// vient d'être appuyée, partagées par tous les écrans du jeu.
// ---------------------------------------------------------------

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

// confirmPressed : Entrée seule (Alt+Entrée est réservé au plein écran).
func confirmPressed() bool {
	return !ebiten.IsKeyPressed(ebiten.KeyAlt) &&
		(inpututil.IsKeyJustPressed(ebiten.KeyEnter) ||
			inpututil.IsKeyJustPressed(ebiten.KeyNumpadEnter))
}

// validatePressed : Entrée OU Espace valident l'option sélectionnée,
// dans tous les menus du jeu.
// Exception volontaire : la saisie du nom (updateNameEntry) utilise
// confirmPressed, car l'Espace n'y doit rien valider.
func validatePressed() bool {
	return confirmPressed() || inpututil.IsKeyJustPressed(ebiten.KeySpace)
}

// Touches de direction dans les menus, en flèches ou en ZQSD / WASD.
// "Pressed" veut dire que la touche vient d'être enfoncée : un appui
// déplace la sélection d'un cran, même si on garde le doigt dessus.
func upPressed() bool {
	return inpututil.IsKeyJustPressed(ebiten.KeyArrowUp) ||
		inpututil.IsKeyJustPressed(ebiten.KeyZ) ||
		inpututil.IsKeyJustPressed(ebiten.KeyW)
}

// downPressed : flèche bas ou S.
func downPressed() bool {
	return inpututil.IsKeyJustPressed(ebiten.KeyArrowDown) ||
		inpututil.IsKeyJustPressed(ebiten.KeyS)
}

// leftPressed : flèche gauche, Q (AZERTY) ou A (QWERTY).
func leftPressed() bool {
	return inpututil.IsKeyJustPressed(ebiten.KeyArrowLeft) ||
		inpututil.IsKeyJustPressed(ebiten.KeyQ) ||
		inpututil.IsKeyJustPressed(ebiten.KeyA)
}

// rightPressed : flèche droite ou D.
func rightPressed() bool {
	return inpututil.IsKeyJustPressed(ebiten.KeyArrowRight) ||
		inpututil.IsKeyJustPressed(ebiten.KeyD)
}

// navigateList fait avancer/reculer une sélection dans une liste
// verticale, avec repli circulaire (après le dernier choix, on revient
// au premier). Utilisée par tous les menus en liste du jeu.
func navigateList(sel *int, n int) {
	if n == 0 {
		return
	}
	if downPressed() {
		*sel = (*sel + 1) % n
	}
	if upPressed() {
		*sel = (*sel - 1 + n) % n
	}
}
