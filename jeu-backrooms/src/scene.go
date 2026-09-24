package main

// ---------------------------------------------------------------
// SALLES — chargement d'une salle (grille de collisions + décor) et
// règles de la carte : murs, portes, arrivée dans la salle suivante.
// Ce fichier ne dessine rien à l'écran, sauf le décor de secours
// quand l'image PNG d'une salle manque.
// ---------------------------------------------------------------

import (
	"bufio"
	"image"
	_ "image/png" // permet de décoder les .png
	"os"
	"path/filepath"
	"slices"
	"strconv"
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

// loadSprite charge assets/<file>. Si le fichier n'existe pas, elle renvoie
// fallback : le visuel de secours dessiné dans le code, pour que le jeu
// tourne même sans les images générées par Python.
func loadSprite(file string, fallback *ebiten.Image) *ebiten.Image {
	img, err := loadPNG(filepath.Join(assetsDir, file))
	if err != nil {
		return fallback
	}
	return img
}

// sceneCache garde les salles déjà chargées : on ne relit pas le fichier
// ni l'image à chaque porte franchie ou à chaque apparition de monstre.
// Une salle n'est jamais modifiée après son chargement, on peut donc la
// partager sans risque.
var sceneCache = map[string]*Scene{}

// LoadScene renvoie la salle demandée, chargée une seule fois puis gardée
// en mémoire (voir readScene).
func LoadScene(name string) (*Scene, error) {
	if s, ok := sceneCache[name]; ok {
		return s, nil
	}
	s, err := readScene(name)
	if err != nil {
		return nil, err
	}
	sceneCache[name] = s
	return s, nil
}

// readScene lit assets/scenes/<name>.txt (collisions) et <name>.png (décor).
// Si le PNG est absent, un décor de secours est dessiné à partir de la grille,
// pour que le jeu tourne quand même.
func readScene(name string) (*Scene, error) {
	s := &Scene{
		Name:  name,
		Exits: make(map[string]string),
		Start: [2]int{-1, -1},
	}
	if err := s.readFile(filepath.Join(assetsDir, "scenes", name+".txt")); err != nil {
		return nil, err
	}
	s.fitGrid()
	s.takeStart()

	if bg, err := loadPNG(filepath.Join(assetsDir, "scenes", name+".png")); err == nil {
		s.Bg = bg
	} else {
		s.Bg = fallbackBackground(s.Grid)
	}
	return s, nil
}

// readFile lit le fichier .txt d'une salle : d'abord l'en-tête (lignes
// "# scene:" et "# exit:"), puis toutes les lignes de la grille.
func (s *Scene) readFile(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	inGrid := false
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := sc.Text()
		if !inGrid && s.readHeaderLine(line) {
			continue
		}
		inGrid = true
		s.Grid = append(s.Grid, []rune(line))
	}
	return sc.Err()
}

// readHeaderLine lit une ligne d'en-tête, "# scene: nom" ou
// "# exit: est -> salle". Renvoie false si la ligne n'en est pas une.
func (s *Scene) readHeaderLine(line string) bool {
	if name, ok := strings.CutPrefix(line, "# scene:"); ok {
		s.Name = strings.TrimSpace(name)
		return true
	}
	exit, ok := strings.CutPrefix(line, "# exit:")
	if !ok {
		return false
	}
	if dir, target, found := strings.Cut(strings.TrimSpace(exit), "->"); found {
		s.Exits[strings.TrimSpace(dir)] = strings.TrimSpace(target)
	}
	return true
}

// fitGrid met la grille à exactement GridH lignes de GridW cases : ce
// qui manque devient du sol '.', ce qui dépasse est coupé.
func (s *Scene) fitGrid() {
	for len(s.Grid) < GridH {
		s.Grid = append(s.Grid, []rune{})
	}
	s.Grid = s.Grid[:GridH]
	for r := range s.Grid {
		for len(s.Grid[r]) < GridW {
			s.Grid[r] = append(s.Grid[r], '.')
		}
		s.Grid[r] = s.Grid[r][:GridW]
	}
}

// takeStart repère le '@' (point de départ du joueur), le note dans
// Start et le remplace par du sol : le joueur est géré à part.
func (s *Scene) takeStart() {
	for r, row := range s.Grid {
		for c, ch := range row {
			if ch == '@' {
				s.Start = [2]int{r, c}
				row[c] = '.'
			}
		}
	}
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

// roomName construit le nom d'une salle à partir de sa ligne et de sa
// colonne sur la carte, par exemple (2, 3) donne "level1_r2c3".
func roomName(row, col int) string {
	return "level1_r" + strconv.Itoa(row) + "c" + strconv.Itoa(col)
}

// tileAt renvoie la case (row, col) de la grille : '#' mur, '.' sol,
// '+' porte. Une case hors de la grille compte comme un mur.
func (s *Scene) tileAt(row, col int) rune {
	if row < 0 || row >= GridH || col < 0 || col >= GridW {
		return '#'
	}
	return s.Grid[row][col]
}

// canMoveInScene teste la boîte de collision "pieds" contre les murs
// d'une scène donnée. Partagée par le joueur (play_input.go) et les monstres
// (monster.go) pour ne pas dupliquer la règle de collision.
func canMoveInScene(scene *Scene, x, y float64) bool {
	left := x + boxOffX
	top := y + boxOffY
	right := left + boxW - 1
	bottom := top + boxH - 1
	if left < 0 || top < 0 || right >= ScreenW || bottom >= ScreenH {
		return false
	}
	corners := [4][2]float64{{left, top}, {right, top}, {left, bottom}, {right, bottom}}
	for _, p := range corners {
		col := int(p[0]) / TileSize
		row := int(p[1]) / TileSize
		if scene.tileAt(row, col) == '#' {
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

// ---------------------------------------------------------------
// ARRIVÉE DANS UNE SALLE — où poser le joueur quand il franchit une
// porte. Ce sont des règles de carte, pas de la boucle de jeu.
// ---------------------------------------------------------------

// reverse donne le bord d'arrivée à partir de la direction empruntée :
// sortir par l'est, c'est entrer par l'ouest de la salle suivante.
var reverse = map[string]string{"nord": "sud", "sud": "nord", "est": "ouest", "ouest": "est"}

// entryPosition place le joueur dans la scène d'arrivée EN CONSERVANT sa
// position le long du mur traversé : on entre à la même hauteur qu'on est sorti.
// Si cette hauteur ne tombe pas en face d'une ouverture, on glisse vers
// l'ouverture la plus proche.
func entryPosition(ns *Scene, traveled string, px, py float64) (float64, float64) {
	side := reverse[traveled]
	nx, ny := stickToWall(side, px, py)

	doors := doorIndices(ns, side)
	if len(doors) == 0 {
		// aucune porte de ce côté : on arrive sur le '@' de la salle s'il existe
		if ns.Start[0] >= 0 {
			return float64(ns.Start[1] * TileSize), float64(ns.Start[0] * TileSize)
		}
		return nx, ny
	}

	if side == "ouest" || side == "est" {
		ny = alignOnDoor(doors, ny+boxOffY+boxH/2, ny)
	} else {
		nx = alignOnDoor(doors, nx+boxOffX+boxW/2, nx)
	}
	return nx, ny
}

// stickToWall pose le joueur juste à l'intérieur du mur d'arrivée ;
// l'autre coordonnée ne change pas.
func stickToWall(side string, x, y float64) (float64, float64) {
	switch side {
	case "ouest":
		x = float64(TileSize)
	case "est":
		x = float64((GridW - 2) * TileSize)
	case "nord":
		y = float64(TileSize)
	case "sud":
		y = float64((GridH - 2) * TileSize)
	}
	return x, y
}

// alignOnDoor garde la position pos si une porte est en face du centre
// des pieds du joueur (feet), sinon renvoie la position de la porte la
// plus proche.
func alignOnDoor(doors []int, feet, pos float64) float64 {
	cur := int(feet) / TileSize
	if slices.Contains(doors, cur) {
		return pos
	}
	return float64(nearestInt(doors, cur) * TileSize)
}

// doorIndices renvoie la position des portes '+' le long d'un bord : un
// numéro de ligne pour ouest/est, un numéro de colonne pour nord/sud.
func doorIndices(s *Scene, side string) []int {
	var out []int
	for i, tile := range borderTiles(s, side) {
		if tile == '+' {
			out = append(out, i)
		}
	}
	return out
}

// borderTiles renvoie, dans l'ordre, les cases d'un bord de la grille.
func borderTiles(s *Scene, side string) []rune {
	switch side {
	case "nord":
		return s.Grid[0]
	case "sud":
		return s.Grid[GridH-1]
	case "ouest", "est":
		col := 0
		if side == "est" {
			col = GridW - 1
		}
		tiles := make([]rune, GridH)
		for r := range tiles {
			tiles[r] = s.Grid[r][col]
		}
		return tiles
	}
	return nil
}

// nearestInt renvoie l'élément de la liste le plus proche de v ; en cas
// d'égalité, le premier trouvé l'emporte.
func nearestInt(list []int, v int) int {
	best, bestD := list[0], 1<<30
	for _, x := range list {
		d := x - v
		if d < 0 {
			d = -d
		}
		if d < bestD {
			best, bestD = x, d
		}
	}
	return best
}
