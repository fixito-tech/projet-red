package main

import "testing"

// Le fichier .txt déclare ses sorties dans l'en-tête ; la salle de
// départ doit mener à la salle suivante vers le sud, comme prévu par
// la carte fournie.
func TestLoadSceneParsesExitsAndStart(t *testing.T) {
	s, err := LoadScene(startScene)
	if err != nil {
		t.Fatalf("scène de départ introuvable : %v", err)
	}
	if s.Name != startRoomName {
		t.Errorf("nom de scène = %q, attendu %q", s.Name, startRoomName)
	}
	if next, ok := s.Exits["sud"]; !ok || next != "level1_r4c2" {
		t.Errorf("sortie sud = %q (ok=%v), attendu level1_r4c2", next, ok)
	}
	if s.Start[0] < 0 {
		t.Error("le point de départ '@' n'a pas été trouvé")
	}
}

// canMoveInScene doit bloquer les murs et autoriser le sol.
func TestCanMoveInSceneBlocksWalls(t *testing.T) {
	s, err := LoadScene("level1_r0c0")
	if err != nil {
		t.Fatalf("scène introuvable : %v", err)
	}
	// (0,0) est un mur d'angle : impossible d'y placer la boîte de collision.
	if canMoveInScene(s, 0, 0) {
		t.Error("le coin (0,0), dans le mur, devrait être bloqué")
	}
	// Le centre de la salle est du sol.
	cx, cy := float64(GridW/2*TileSize), float64(GridH/2*TileSize)
	if !canMoveInScene(s, cx, cy) {
		t.Error("le centre de la salle, sur le sol, devrait être accessible")
	}
}

// Toutes les salles de la carte 5x5 doivent pouvoir être chargées et
// avoir une grille complète GridH x GridW (le jeu doit s'adapter à
// n'importe quelle carte, mais celle fournie doit être valide).
func TestAllRoomsLoadWithFullGrid(t *testing.T) {
	for _, name := range AllRoomNames() {
		s, err := LoadScene(name)
		if err != nil {
			t.Fatalf("%s : %v", name, err)
		}
		if len(s.Grid) != GridH {
			t.Fatalf("%s : %d lignes, attendu %d", name, len(s.Grid), GridH)
		}
		for r, row := range s.Grid {
			if len(row) != GridW {
				t.Fatalf("%s : ligne %d = %d colonnes, attendu %d", name, r, len(row), GridW)
			}
		}
	}
}

// entryPosition doit toujours renvoyer une position dans les limites
// de l'écran, quelle que soit la salle d'arrivée.
func TestEntryPositionStaysInBounds(t *testing.T) {
	ns, err := LoadScene("level1_r4c2")
	if err != nil {
		t.Fatalf("scène introuvable : %v", err)
	}
	nx, ny := entryPosition(ns, "sud", float64(12*TileSize), float64(0))
	if nx < 0 || nx >= ScreenW || ny < 0 || ny >= ScreenH {
		t.Errorf("position d'entrée hors écran : (%.0f, %.0f)", nx, ny)
	}
}

// La salle du boss est bien en haut au milieu de la carte 5x5.
func TestBossRoomIsTopMiddle(t *testing.T) {
	if bossRoomName != "level1_r0c2" {
		t.Errorf("bossRoomName = %q, attendu level1_r0c2 (rangée 0, colonne du milieu)", bossRoomName)
	}
}
