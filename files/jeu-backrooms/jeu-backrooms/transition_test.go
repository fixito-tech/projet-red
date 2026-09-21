package main

import "testing"

// buildScene fabrique une scène avec un mur de contour et une ouverture
// sur un bord, de la case lo à la case hi (inclus).
func buildScene(side string, lo, hi int) *Scene {
	g := make([][]rune, GridH)
	for r := 0; r < GridH; r++ {
		g[r] = make([]rune, GridW)
		for c := 0; c < GridW; c++ {
			if r == 0 || r == GridH-1 || c == 0 || c == GridW-1 {
				g[r][c] = '#'
			} else {
				g[r][c] = '.'
			}
		}
	}
	for i := lo; i <= hi; i++ {
		switch side {
		case "ouest":
			g[i][0] = '+'
		case "est":
			g[i][GridW-1] = '+'
		case "nord":
			g[0][i] = '+'
		case "sud":
			g[GridH-1][i] = '+'
		}
	}
	return &Scene{Grid: g, Start: [2]int{-1, -1}, Exits: map[string]string{}}
}

// On sort vers l'est à la hauteur d'une case donnée : on doit entrer
// dans la scène suivante exactement à la même hauteur.
func TestHauteurConserveeVersEst(t *testing.T) {
	ns := buildScene("ouest", 4, 10) // ouverture large côté ouest
	for _, row := range []int{4, 6, 9, 10} {
		py := float64(row * TileSize)
		nx, ny := entryPosition(ns, "est", 700, py)
		if ny != py {
			t.Errorf("sortie ligne %d : hauteur %v attendue, obtenue %v", row, py, ny)
		}
		if nx != float64(TileSize) {
			t.Errorf("ligne %d : x attendu %d, obtenu %v", row, TileSize, nx)
		}
	}
}

// Même chose vers l'ouest.
func TestHauteurConserveeVersOuest(t *testing.T) {
	ns := buildScene("est", 2, 12)
	py := float64(7 * TileSize)
	nx, ny := entryPosition(ns, "ouest", 40, py)
	if ny != py {
		t.Errorf("hauteur %v attendue, obtenue %v", py, ny)
	}
	if nx != float64((GridW-2)*TileSize) {
		t.Errorf("x attendu %d, obtenu %v", (GridW-2)*TileSize, nx)
	}
}

// Vers le sud/nord, c'est la colonne qui doit être conservée.
func TestColonneConserveeVersSud(t *testing.T) {
	ns := buildScene("nord", 3, 20)
	px := float64(15 * TileSize)
	nx, ny := entryPosition(ns, "sud", px, 440)
	if nx != px {
		t.Errorf("colonne %v attendue, obtenue %v", px, nx)
	}
	if ny != float64(TileSize) {
		t.Errorf("y attendu %d, obtenu %v", TileSize, ny)
	}
}

// Si on sort en face d'un mur plein, on doit glisser vers l'ouverture
// la plus proche au lieu d'atterrir dans le décor.
func TestGlisseVersOuvertureLaPlusProche(t *testing.T) {
	ns := buildScene("ouest", 6, 8) // ouverture étroite lignes 6..8
	_, ny := entryPosition(ns, "est", 700, float64(12*TileSize))
	if ny != float64(8*TileSize) {
		t.Errorf("attendu ligne 8 (%v), obtenu %v", float64(8*TileSize), ny)
	}
	_, ny2 := entryPosition(ns, "est", 700, float64(1*TileSize))
	if ny2 != float64(6*TileSize) {
		t.Errorf("attendu ligne 6 (%v), obtenu %v", float64(6*TileSize), ny2)
	}
}
