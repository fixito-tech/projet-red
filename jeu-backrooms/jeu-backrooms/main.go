package main

import (
	"image"
	"image/color"
	"log"
	"path/filepath"
	"strconv"
	"unicode"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

// ---------------------------------------------------------------
// RÉGLAGES — tout ce que tu peux vouloir changer est ici
// ---------------------------------------------------------------
const (
	assetsDir  = "assets"        // dossier des ressources
	startScene = "level1_start"  // scène de départ

	TileSize = 32 // taille d'une case en pixels
	GridW    = 25 // nombre de cases en largeur
	GridH    = 15 // nombre de cases en hauteur

	ScreenW = TileSize * GridW // 800
	ScreenH = TileSize * GridH // 480

	startFullscreen = false // true pour démarrer directement en plein écran

	playerSpeed = 2.0 // pixels par frame
	animSpeed   = 8   // ticks entre deux images d'animation
)

// Boîte de collision du joueur (plus petite que le sprite : les "pieds").
const (
	boxOffX = 8
	boxOffY = 18
	boxW    = 16
	boxH    = 12
)

// Directions — correspondent aux LIGNES du spritesheet.
const (
	DirDown  = 0
	DirLeft  = 1
	DirRight = 2
	DirUp    = 3
)

var dirName = map[int]string{DirDown: "sud", DirLeft: "ouest", DirRight: "est", DirUp: "nord"}
var reverse = map[string]string{"nord": "sud", "sud": "nord", "est": "ouest", "ouest": "est"}

// ---------------------------------------------------------------
// ÉTAT DU JEU
// ---------------------------------------------------------------
type State int

const (
	StateMenu State = iota
	StateCreate // saisie du nom puis choix de la classe
	StatePlay
)

type Game struct {
	state State

	// menu
	menuBg    *ebiten.Image
	menuIndex int

	// scène + joueur
	scene  *Scene
	px, py float64 // coin haut-gauche du sprite
	dir    int
	frame  int
	tick   int

	// sprite
	sheet   *ebiten.Image
	frameW  int
	frameH  int
	nFrames int

	showDebug bool
	textCache map[string]*ebiten.Image

	// personnage et interface
	hero       *Character
	invOpen    bool
	invSel     int
	msg        string
	msgTimer   int
	createStep int    // 0 = nom, 1 = classe
	nameBuf    []rune // nom en cours de saisie
	classSel   int
	uiTick     int
}

// ---------------------------------------------------------------
// MISE À JOUR
// ---------------------------------------------------------------
func (g *Game) Update() error {
	// Plein écran : F11 ou Alt+Entrée, disponible dans le menu comme en jeu.
	if inpututil.IsKeyJustPressed(ebiten.KeyF11) ||
		(inpututil.IsKeyJustPressed(ebiten.KeyEnter) && ebiten.IsKeyPressed(ebiten.KeyAlt)) {
		ebiten.SetFullscreen(!ebiten.IsFullscreen())
	}

	g.uiTick++
	if g.msgTimer > 0 {
		g.msgTimer--
		if g.msgTimer == 0 {
			g.msg = ""
		}
	}

	switch g.state {
	case StateMenu:
		return g.updateMenu()
	case StateCreate:
		return g.updateCreate()
	case StatePlay:
		return g.updatePlay()
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

func confirmPressed() bool {
	return !ebiten.IsKeyPressed(ebiten.KeyAlt) &&
		(inpututil.IsKeyJustPressed(ebiten.KeyEnter) ||
			inpututil.IsKeyJustPressed(ebiten.KeyNumpadEnter))
}

func (g *Game) updateMenu() error {
	items := g.menuEntries()
	if g.menuIndex >= len(items) {
		g.menuIndex = 0
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyArrowDown) || inpututil.IsKeyJustPressed(ebiten.KeyS) {
		g.menuIndex = (g.menuIndex + 1) % len(items)
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyArrowUp) || inpututil.IsKeyJustPressed(ebiten.KeyZ) || inpututil.IsKeyJustPressed(ebiten.KeyW) {
		g.menuIndex = (g.menuIndex - 1 + len(items)) % len(items)
	}
	if confirmPressed() || inpututil.IsKeyJustPressed(ebiten.KeySpace) {
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

// updateCreate : écran de création du personnage (tâche 11).
func (g *Game) updateCreate() error {
	if g.createStep == 0 {
		// saisie du nom : on ne garde que les lettres
		for _, r := range ebiten.AppendInputChars(nil) {
			if unicode.IsLetter(r) && len(g.nameBuf) < 14 {
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
		return nil
	}

	// choix de la classe
	if inpututil.IsKeyJustPressed(ebiten.KeyArrowLeft) || inpututil.IsKeyJustPressed(ebiten.KeyQ) || inpututil.IsKeyJustPressed(ebiten.KeyA) {
		g.classSel = (g.classSel - 1 + len(Classes)) % len(Classes)
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyArrowRight) || inpututil.IsKeyJustPressed(ebiten.KeyD) {
		g.classSel = (g.classSel + 1) % len(Classes)
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		g.createStep = 0
	}
	if confirmPressed() || inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		g.hero = NewCharacter(FormatNom(string(g.nameBuf)), Classes[g.classSel])
		g.startNewGame()
		g.state = StatePlay
	}
	return nil
}

// startNewGame remet le joueur dans la scène de départ.
func (g *Game) startNewGame() {
	if sc, err := LoadScene(startScene); err == nil {
		g.scene = sc
	}
	if g.scene.Start[0] >= 0 {
		g.px = float64(g.scene.Start[1] * TileSize)
		g.py = float64(g.scene.Start[0] * TileSize)
	} else {
		g.px = float64(GridW / 2 * TileSize)
		g.py = float64(GridH / 2 * TileSize)
	}
	g.dir, g.frame, g.tick = DirDown, 0, 0
	g.invOpen, g.invSel = false, 0
}

func (g *Game) toast(m string) {
	g.msg = m
	g.msgTimer = 120 // 2 secondes
}

// updateInventory : navigation dans la grille quand l'inventaire est ouvert.
func (g *Game) updateInventory() {
	h := g.hero
	cols, _, _ := invLayout(h.Capacite)
	if inpututil.IsKeyJustPressed(ebiten.KeyE) || inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		g.invOpen = false
		return
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyArrowRight) || inpututil.IsKeyJustPressed(ebiten.KeyD) {
		g.invSel++
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyArrowLeft) || inpututil.IsKeyJustPressed(ebiten.KeyQ) || inpututil.IsKeyJustPressed(ebiten.KeyA) {
		g.invSel--
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyArrowDown) || inpututil.IsKeyJustPressed(ebiten.KeyS) {
		g.invSel += cols
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyArrowUp) || inpututil.IsKeyJustPressed(ebiten.KeyZ) || inpututil.IsKeyJustPressed(ebiten.KeyW) {
		g.invSel -= cols
	}
	if g.invSel < 0 {
		g.invSel = 0
	}
	if g.invSel >= h.Capacite {
		g.invSel = h.Capacite - 1
	}
	if confirmPressed() || inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		if m := h.UseItem(g.invSel); m != "" {
			g.toast(m)
		}
	}
}

// Touches de test pour la démo (à retirer ou garder pour l'oral).
var testItems = []string{"Morceau de moquette", "Tuyau rouillé", "Néon cassé", "Barre de fer", "Almond Water"}

func (g *Game) updateDebugKeys() {
	h := g.hero
	if inpututil.IsKeyJustPressed(ebiten.KeyF2) { // ajoute un objet
		item := testItems[g.uiTick%len(testItems)]
		if h.AddItem(item) {
			g.toast("Ramassé : " + item)
		} else {
			g.toast("Inventaire plein !")
		}
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyF3) { // prend des dégâts
		h.PV -= 15
		if h.PV < 0 {
			h.PV = 0
		}
		h.Energie -= 10
		if h.Energie < 0 {
			h.Energie = 0
		}
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyF4) { // agrandit l'inventaire
		if h.UpgradeInventory() {
			g.toast("Inventaire agrandi : " + strconv.Itoa(h.Capacite) + " places")
		} else {
			g.toast("Amélioration maximale atteinte")
		}
	}
}

func (g *Game) updatePlay() error {
	g.updateDebugKeys()

	// Inventaire ouvert : il capte les touches et le joueur ne bouge plus.
	if g.invOpen {
		g.updateInventory()
		g.frame, g.tick = 0, 0
		return nil
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyE) {
		g.invOpen = true
		return nil
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		g.state = StateMenu
		return nil
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyF1) {
		g.showDebug = !g.showDebug
	}

	dx, dy := 0.0, 0.0
	if ebiten.IsKeyPressed(ebiten.KeyArrowLeft) || ebiten.IsKeyPressed(ebiten.KeyQ) || ebiten.IsKeyPressed(ebiten.KeyA) {
		dx -= playerSpeed
		g.dir = DirLeft
	}
	if ebiten.IsKeyPressed(ebiten.KeyArrowRight) || ebiten.IsKeyPressed(ebiten.KeyD) {
		dx += playerSpeed
		g.dir = DirRight
	}
	if ebiten.IsKeyPressed(ebiten.KeyArrowUp) || ebiten.IsKeyPressed(ebiten.KeyZ) || ebiten.IsKeyPressed(ebiten.KeyW) {
		dy -= playerSpeed
		g.dir = DirUp
	}
	if ebiten.IsKeyPressed(ebiten.KeyArrowDown) || ebiten.IsKeyPressed(ebiten.KeyS) {
		dy += playerSpeed
		g.dir = DirDown
	}

	moving := dx != 0 || dy != 0

	// Déplacement axe par axe : on glisse le long des murs au lieu de se bloquer.
	if dx != 0 && g.canMove(g.px+dx, g.py) {
		g.px += dx
	}
	if dy != 0 && g.canMove(g.px, g.py+dy) {
		g.py += dy
	}

	// Animation
	if moving {
		g.tick++
		if g.tick >= animSpeed {
			g.tick = 0
			g.frame = (g.frame + 1) % g.nFrames
		}
	} else {
		g.frame = 0
		g.tick = 0
	}

	g.checkTransition()
	return nil
}

// canMove teste la boîte de collision aux quatre coins contre les murs '#'.
func (g *Game) canMove(x, y float64) bool {
	left := x + boxOffX
	top := y + boxOffY
	right := left + boxW - 1
	bottom := top + boxH - 1

	if left < 0 || top < 0 || right >= ScreenW || bottom >= ScreenH {
		return false
	}
	corners := [4][2]float64{{left, top}, {right, top}, {left, bottom}, {right, bottom}}
	for _, p := range corners {
		c := int(p[0]) / TileSize
		r := int(p[1]) / TileSize
		if r < 0 || r >= GridH || c < 0 || c >= GridW {
			return false
		}
		if g.scene.Grid[r][c] == '#' {
			return false
		}
	}
	return true
}

// checkTransition : si le joueur est sur une case '+', on change de scène.
func (g *Game) checkTransition() {
	cx := g.px + boxOffX + boxW/2
	cy := g.py + boxOffY + boxH/2
	col := int(cx) / TileSize
	row := int(cy) / TileSize
	if row < 0 || row >= GridH || col < 0 || col >= GridW {
		return
	}
	if g.scene.Grid[row][col] != '+' {
		return
	}

	// Direction de sortie : selon le bord touché, sinon selon l'orientation.
	traveled := ""
	switch {
	case col == 0:
		traveled = "ouest"
	case col == GridW-1:
		traveled = "est"
	case row == 0:
		traveled = "nord"
	case row == GridH-1:
		traveled = "sud"
	default:
		traveled = dirName[g.dir]
	}

	next, ok := g.scene.Exits[traveled]
	if !ok {
		return // sortie non reliée dans l'en-tête du .txt
	}
	g.goTo(next, traveled)
}

// goTo charge la scène suivante et y place le joueur.
func (g *Game) goTo(name, traveled string) {
	ns, err := LoadScene(name)
	if err != nil {
		log.Printf("scène introuvable %q : %v", name, err)
		return
	}
	nx, ny := entryPosition(ns, traveled, g.px, g.py)

	g.scene = ns
	g.px, g.py = nx, ny

	// Sécurité : si on atterrit dans un mur, on décale vers l'intérieur.
	for i := 0; i < GridW && !g.canMove(g.px, g.py); i++ {
		switch traveled {
		case "est":
			g.px += TileSize
		case "ouest":
			g.px -= TileSize
		case "sud":
			g.py += TileSize
		case "nord":
			g.py -= TileSize
		}
		if g.px < 0 || g.py < 0 || g.px > ScreenW || g.py > ScreenH {
			g.px, g.py = nx, ny
			break
		}
	}
}

// doorIndices renvoie les index des cases '+' le long d'un bord donné.
func doorIndices(s *Scene, side string) []int {
	var out []int
	switch side {
	case "ouest":
		for r := 0; r < GridH; r++ {
			if s.Grid[r][0] == '+' {
				out = append(out, r)
			}
		}
	case "est":
		for r := 0; r < GridH; r++ {
			if s.Grid[r][GridW-1] == '+' {
				out = append(out, r)
			}
		}
	case "nord":
		for c := 0; c < GridW; c++ {
			if s.Grid[0][c] == '+' {
				out = append(out, c)
			}
		}
	case "sud":
		for c := 0; c < GridW; c++ {
			if s.Grid[GridH-1][c] == '+' {
				out = append(out, c)
			}
		}
	}
	return out
}

func containsInt(list []int, v int) bool {
	for _, x := range list {
		if x == v {
			return true
		}
	}
	return false
}

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

// entryPosition place le joueur dans la scène d'arrivée EN CONSERVANT sa
// position le long du mur traversé : on entre à la même hauteur qu'on est sorti.
// Si cette hauteur ne tombe pas en face d'une ouverture, on glisse vers
// l'ouverture la plus proche.
func entryPosition(ns *Scene, traveled string, px, py float64) (float64, float64) {
	nx, ny := px, py
	side := reverse[traveled]

	// On se colle juste à l'intérieur du mur d'arrivée, l'autre axe est conservé.
	switch side {
	case "ouest":
		nx = float64(TileSize)
	case "est":
		nx = float64((GridW - 2) * TileSize)
	case "nord":
		ny = float64(TileSize)
	case "sud":
		ny = float64((GridH - 2) * TileSize)
	}

	idxs := doorIndices(ns, side)
	if len(idxs) == 0 {
		if ns.Start[0] >= 0 {
			return float64(ns.Start[1] * TileSize), float64(ns.Start[0] * TileSize)
		}
		return nx, ny
	}

	if side == "ouest" || side == "est" {
		cur := int(ny+boxOffY+boxH/2) / TileSize
		if !containsInt(idxs, cur) {
			ny = float64(nearestInt(idxs, cur) * TileSize)
		}
	} else {
		cur := int(nx+boxOffX+boxW/2) / TileSize
		if !containsInt(idxs, cur) {
			nx = float64(nearestInt(idxs, cur) * TileSize)
		}
	}
	return nx, ny
}

// ---------------------------------------------------------------
// AFFICHAGE
// ---------------------------------------------------------------
func (g *Game) Draw(screen *ebiten.Image) {
	switch g.state {
	case StateMenu:
		g.drawMenu(screen)
	case StateCreate:
		if g.createStep == 0 {
			drawNameStep(screen, string(g.nameBuf), g.uiTick)
		} else {
			drawClassStep(screen, g)
		}
	case StatePlay:
		g.drawPlay(screen)
	}
}

func (g *Game) drawMenu(screen *ebiten.Image) {
	if g.menuBg != nil {
		op := &ebiten.DrawImageOptions{}
		w, h := g.menuBg.Bounds().Dx(), g.menuBg.Bounds().Dy()
		op.GeoM.Scale(float64(ScreenW)/float64(w), float64(ScreenH)/float64(h))
		screen.DrawImage(g.menuBg, op)
	} else {
		drawCreateBackground(screen)
		txt(screen, "THE BACKROOMS", ScreenW/2, 110, 44, true, colGold, alignCenter)
		txt(screen, "Projet RED", ScreenW/2, 170, 14, false, colMuted, alignCenter)
	}

	for i, item := range g.menuEntries() {
		y := 250.0 + float64(i)*46
		c := color.Color(colMuted)
		if i == g.menuIndex {
			c = colGold
			txt(screen, "›", ScreenW/2-110, y-3, 26, true, colGold, alignCenter)
		}
		txt(screen, item, ScreenW/2, y, 22, true, c, alignCenter)
	}
	txt(screen, "Flèches ou ZQSD  ·  Entrée pour valider  ·  F11 : plein écran",
		ScreenW/2, 440, 12, false, colMuted, alignCenter)
}

func (g *Game) drawPlay(screen *ebiten.Image) {
	screen.DrawImage(g.scene.Bg, nil)

	// Le sprite peut être plus grand qu'une case : il déborde vers le haut
	// et sur les côtés, mais son "empreinte au sol" reste de 32x32 pixels.
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(
		g.px-float64(g.frameW-TileSize)/2,
		g.py-float64(g.frameH-TileSize),
	)
	screen.DrawImage(g.playerFrame(), op)

	// Le HUD passe APRÈS l'inventaire pour rester lisible par-dessus le voile :
	// on voit la barre de vie remonter quand on boit une Almond Water.
	if g.invOpen {
		drawInventory(screen, g.hero, g.invSel)
	}
	drawHUD(screen, g.hero, !g.invOpen)
	drawToast(screen, g.msg)

	if g.showDebug {
		g.drawText(screen, "scene: "+g.scene.Name, 8, 100, 1)
		y := 116.0
		for dir, target := range g.scene.Exits {
			g.drawText(screen, dir+" -> "+target, 8, y, 1)
			y += 14
		}
	}
}

// playerFrame renvoie l'image du sprite correspondant à la direction + l'anim.
func (g *Game) playerFrame() *ebiten.Image {
	sx := g.frame * g.frameW
	sy := g.dir * g.frameH
	r := image.Rect(sx, sy, sx+g.frameW, sy+g.frameH)
	return g.sheet.SubImage(r).(*ebiten.Image)
}

// drawText affiche du texte agrandi (police intégrée d'Ebiten, aucune dépendance).
func (g *Game) drawText(dst *ebiten.Image, s string, x, y float64, scale float64) {
	img, ok := g.textCache[s]
	if !ok {
		img = ebiten.NewImage(len(s)*6+8, 16)
		ebitenutil.DebugPrintAt(img, s, 0, 0)
		g.textCache[s] = img
	}
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(scale, scale)
	op.GeoM.Translate(x, y)
	dst.DrawImage(img, op)
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return ScreenW, ScreenH
}

// ---------------------------------------------------------------
// SPRITE DE SECOURS (si assets/player.png est absent)
// ---------------------------------------------------------------
func placeholderSheet() *ebiten.Image {
	const fw, fh, cols, rows = 32, 32, 4, 4
	img := image.NewRGBA(image.Rect(0, 0, fw*cols, fh*rows))

	body := color.RGBA{0x2b, 0x4a, 0x8b, 0xff}
	skin := color.RGBA{0xf0, 0xcb, 0xa0, 0xff}

	fillRect := func(x0, y0, w, h int, c color.RGBA) {
		for y := y0; y < y0+h; y++ {
			for x := x0; x < x0+w; x++ {
				img.Set(x, y, c)
			}
		}
	}

	for row := 0; row < rows; row++ {
		for col := 0; col < cols; col++ {
			ox, oy := col*fw, row*fh
			fillRect(ox+11, oy+4, 10, 9, skin)    // tête
			fillRect(ox+9, oy+13, 14, 12, body)   // corps
			// jambes animées : elles alternent selon la frame
			legShift := 0
			if col%2 == 1 {
				legShift = 2
			}
			fillRect(ox+10, oy+25, 4, 5-legShift, body)
			fillRect(ox+18, oy+25, 4, 3+legShift, body)
		}
	}
	return ebiten.NewImageFromImage(img)
}

// ---------------------------------------------------------------
// DÉMARRAGE
// ---------------------------------------------------------------
func main() {
	scene, err := LoadScene(startScene)
	if err != nil {
		log.Fatalf("impossible de charger la scène de départ (%s) : %v",
			filepath.Join(assetsDir, "scenes", startScene+".txt"), err)
	}

	g := &Game{
		state:     StateMenu,
		scene:     scene,
		textCache: make(map[string]*ebiten.Image),
	}

	// Spritesheet : 4 colonnes (animation) x 4 lignes (bas, gauche, droite, haut).
	if sheet, err := loadPNG(filepath.Join(assetsDir, "player.png")); err == nil {
		g.sheet = sheet
	} else {
		g.sheet = placeholderSheet()
	}
	g.nFrames = 4
	g.frameW = g.sheet.Bounds().Dx() / g.nFrames
	g.frameH = g.sheet.Bounds().Dy() / 4

	// Image de fond du menu (optionnelle).
	if mb, err := loadPNG(filepath.Join(assetsDir, "menu.png")); err == nil {
		g.menuBg = mb
	}

	// Position de départ du joueur.
	if scene.Start[0] >= 0 {
		g.px = float64(scene.Start[1] * TileSize)
		g.py = float64(scene.Start[0] * TileSize)
	} else {
		g.px = float64(GridW / 2 * TileSize)
		g.py = float64(GridH / 2 * TileSize)
	}

	// Fenêtre deux fois plus grande que la résolution interne : plus
	// confortable, et l'image reste nette (agrandissement entier).
	ebiten.SetWindowSize(ScreenW*2, ScreenH*2)
	ebiten.SetWindowTitle("Projet RED — The Backrooms")
	// Autorise le redimensionnement : sans ça, le bouton "agrandir" de la
	// barre de titre est grisé.
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	if startFullscreen {
		ebiten.SetFullscreen(true)
	}
	if err := ebiten.RunGame(g); err != nil {
		log.Fatal(err)
	}
}
