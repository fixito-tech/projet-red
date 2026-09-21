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
