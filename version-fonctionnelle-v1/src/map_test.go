package main

import "testing"

func TestMapIsRectangular(t *testing.T) {
	for y, row := range worldMap {
		if len(row) != len(worldMap[0]) {
			t.Errorf("la ligne %d a une largeur différente de la première ligne", y)
		}
	}
}

func TestMapStartIsWalkable(t *testing.T) {
	if tileAt(startX, startY) != '.' {
		t.Errorf("la case de départ doit être un sol, got %q", tileAt(startX, startY))
	}
}

func TestMapPlacesAreReachable(t *testing.T) {
	type pos struct{ x, y int }
	seen := map[pos]bool{{startX, startY}: true}
	queue := []pos{{startX, startY}}
	for len(queue) > 0 {
		p := queue[0]
		queue = queue[1:]
		for _, d := range []pos{{0, -1}, {0, 1}, {-1, 0}, {1, 0}} {
			n := pos{p.x + d.x, p.y + d.y}
			if seen[n] || tileAt(n.x, n.y) == '#' {
				continue
			}
			seen[n] = true
			queue = append(queue, n)
		}
	}

	for y, row := range worldMap {
		for x := 0; x < len(row); x++ {
			tile := row[x]
			if (tile == 'T' || tile == 'B' || tile == 'H') && !seen[pos{x, y}] {
				t.Errorf("la case %q en (%d,%d) est inaccessible", tile, x, y)
			}
		}
	}
}

func TestTileAtOutsideIsWall(t *testing.T) {
	if tileAt(-1, 0) != '#' || tileAt(0, -1) != '#' || tileAt(100, 100) != '#' {
		t.Error("hors de la carte, la case doit être un mur")
	}
}

func TestMoveOffset(t *testing.T) {
	if dx, dy, ok := moveOffset('d'); !ok || dx != 1 || dy != 0 {
		t.Errorf("d doit aller à droite, got %d,%d,%v", dx, dy, ok)
	}
	if _, _, ok := moveOffset('x'); ok {
		t.Error("x n'est pas une touche de déplacement")
	}
}
