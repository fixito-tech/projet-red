package main

// ---------------------------------------------------------------
// POINT D'ENTRÉE — charge la salle de départ, prépare le jeu, puis
// confie la boucle à Ebiten (voir game.go). Chaque écran a son
// fichier *_input.go (clavier) et *_ui.go (dessin) ; les règles du
// jeu sont dans player.go, inventory.go, combat.go, economy.go…
// ---------------------------------------------------------------

import (
	"log"
	"path/filepath"

	"github.com/hajimehoshi/ebiten/v2"
)

// main charge la salle de départ, prépare le jeu et lance la boucle d'Ebiten.
func main() {
	// La salle de départ est chargée tout de suite : si elle manque, on
	// arrête avec un message clair plutôt que de planter plus tard.
	scene, err := LoadScene(startScene)
	if err != nil {
		log.Fatalf("impossible de charger la scène de départ (%s) : %v",
			filepath.Join(assetsDir, "scenes", startScene+".txt"), err)
	}

	g := &Game{
		state: StateMenu,
		scene: scene,
	}
	g.loadImages()
	setupWindow()

	if err := ebiten.RunGame(g); err != nil {
		log.Fatal(err)
	}
}
