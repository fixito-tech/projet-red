package main

import (
	"bufio"
	"image"
	"image/color"
	_ "image/png" // permet de décoder les .png
	"os"
	"path/filepath"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
)

// Scene = un "écran" du jeu : une image de décor + une grille de collisions.
type Scene struct {
	Name  string
	Exits map[string]string // "est" -> "level1_hall"
	Grid  [][]rune          // GridH lignes x GridW colonnes
	Start [2]int            // {ligne, colonne} du '@' du fichier, ou {-1,-1}
	Bg    *ebiten.Image     // le décor PNG
}

// loadPNG ouvre une image PNG et la convertit pour Ebiten.
func loadPNG(path string) (*ebiten.Image, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	img, _, err := image.Decode(f)
	if err != nil {
		return nil, err
	}
	return ebiten.NewImageFromImage(img), nil
}

// LoadScene charge assets/scenes/<name>.txt (collisions) et <name>.png (décor).
// Si le PNG est absent, un décor de secours est dessiné à partir de la grille,
// pour que le jeu tourne quand même.
func LoadScene(name string) (*Scene, error) {
	s := &Scene{
		Name:  name,
		Exits: make(map[string]string),
		Start: [2]int{-1, -1},
	}

	txtPath := filepath.Join(assetsDir, "scenes", name+".txt")
	f, err := os.Open(txtPath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	inGrid := false
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := sc.Text()
		if !inGrid {
			if strings.HasPrefix(line, "# scene:") {
				s.Name = strings.TrimSpace(line[len("# scene:"):])
				continue
			}
			if strings.HasPrefix(line, "# exit:") {
				body := strings.TrimSpace(line[len("# exit:"):])
				if parts := strings.SplitN(body, "->", 2); len(parts) == 2 {
					s.Exits[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
				}
				continue
			}
			inGrid = true
		}
		s.Grid = append(s.Grid, []rune(line))
	}
	if err := sc.Err(); err != nil {
		return nil, err
	}

	// Normalise la grille à exactement GridH x GridW.
	for len(s.Grid) < GridH {
		s.Grid = append(s.Grid, []rune{})
	}
	s.Grid = s.Grid[:GridH]
	for r := range s.Grid {
		for len(s.Grid[r]) < GridW {
			s.Grid[r] = append(s.Grid[r], '.')
		}
		s.Grid[r] = s.Grid[r][:GridW]
		for c, ch := range s.Grid[r] {
			if ch == '@' {
				s.Start = [2]int{r, c}
				s.Grid[r][c] = '.' // le joueur est géré à part
			}
		}
	}

	if bg, err := loadPNG(filepath.Join(assetsDir, "scenes", name+".png")); err == nil {
		s.Bg = bg
	} else {
		s.Bg = fallbackBackground(s.Grid)
	}
	return s, nil
}

// AllRoomNames énumère les 25 salles de la carte 5x5 (level1_r0c0 à
// level1_r4c4), indépendamment du fichier "level1_start" qui n'est qu'un
// alias de la salle de départ (voir startRoomName).
func AllRoomNames() []string {
	names := make([]string, 0, LevelRows*LevelCols)
	for r := 0; r < LevelRows; r++ {
		for c := 0; c < LevelCols; c++ {
			names = append(names, roomName(r, c))
		}
	}
	return names
}

func roomName(row, col int) string {
	return "level1_r" + itoa(row) + "c" + itoa(col)
}

// itoa évite d'importer strconv juste pour deux chiffres 0-4.
func itoa(n int) string {
	if n < 0 || n > 9 {
		return "0"
	}
	return string(rune('0' + n))
}

// boxCorners renvoie les quatre coins d'une boîte de collision alignée
// sur les pieds (mêmes dimensions que celle du joueur), utile pour les
// monstres qui partagent la même règle de collision.
func boxCorners(x, y float64) [4][2]float64 {
	left := x + boxOffX
	top := y + boxOffY
	right := left + boxW - 1
	bottom := top + boxH - 1
	return [4][2]float64{{left, top}, {right, top}, {left, bottom}, {right, bottom}}
}

// canMoveInScene teste la boîte de collision "pieds" contre les murs
// d'une scène donnée. Partagée par le joueur (main.go) et les monstres
// (monster.go) pour ne pas dupliquer la règle de collision.
func canMoveInScene(scene *Scene, x, y float64) bool {
	left := x + boxOffX
	top := y + boxOffY
	right := left + boxW - 1
	bottom := top + boxH - 1
	if left < 0 || top < 0 || right >= ScreenW || bottom >= ScreenH {
		return false
	}
	for _, p := range boxCorners(x, y) {
		col := int(p[0]) / TileSize
		row := int(p[1]) / TileSize
		if row < 0 || row >= GridH || col < 0 || col >= GridW {
			return false
		}
		if scene.Grid[row][col] == '#' {
			return false
		}
	}
	return true
}

// boxesOverlap teste le contact entre deux boîtes de collision "pieds"
// situées en (x1,y1) et (x2,y2) : utilisé pour le contact joueur/monstre
// et le ramassage du loot au sol.
func boxesOverlap(x1, y1, x2, y2 float64) bool {
	l1, t1 := x1+boxOffX, y1+boxOffY
	l2, t2 := x2+boxOffX, y2+boxOffY
	return l1 < l2+boxW && l2 < l1+boxW && t1 < t2+boxH && t2 < t1+boxH
}

// fallbackBackground dessine un décor jaune "backrooms" basique depuis la grille,
// utilisé quand le PNG de la scène n'existe pas encore.
func fallbackBackground(grid [][]rune) *ebiten.Image {
	floor := color.RGBA{0xc9, 0xbb, 0x3d, 0xff}
	wall := color.RGBA{0xa8, 0x99, 0x2a, 0xff}
	door := color.RGBA{0x6b, 0x5e, 0x12, 0xff}

	img := image.NewRGBA(image.Rect(0, 0, ScreenW, ScreenH))
	for r := 0; r < GridH; r++ {
		for c := 0; c < GridW; c++ {
			col := floor
			switch grid[r][c] {
			case '#':
				col = wall
			case '+':
				col = door
			}
			for y := r * TileSize; y < (r+1)*TileSize; y++ {
				for x := c * TileSize; x < (c+1)*TileSize; x++ {
					img.Set(x, y, col)
				}
			}
		}
	}
	return ebiten.NewImageFromImage(img)
}
