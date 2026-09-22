package main

// ---------------------------------------------------------------
// INTERFACE — tout ce qui est dessiné par-dessus le jeu :
// barres de vie/énergie, inventaire, création du personnage.
// ---------------------------------------------------------------

import (
	"bytes"
	"image"
	"image/color"
	"log"
	"strconv"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"golang.org/x/image/font/gofont/gobold"
	"golang.org/x/image/font/gofont/goregular"
)

// ---------- couleurs ----------
var (
	colText    = color.RGBA{0xf2, 0xec, 0xd4, 0xff}
	colMuted   = color.RGBA{0xb3, 0xab, 0x88, 0xff}
	colGold    = color.RGBA{0xe0, 0xb8, 0x2e, 0xff}
	colPanel   = color.RGBA{0x24, 0x21, 0x12, 0xf0}
	colPanelLt = color.RGBA{0x33, 0x2f, 0x1b, 0xff}
	colBorder  = color.RGBA{0x6b, 0x61, 0x33, 0xff}
	colSlot    = color.RGBA{0x17, 0x15, 0x0b, 0xff}
	colHP      = color.RGBA{0xd6, 0x45, 0x45, 0xff}
	colHPBack  = color.RGBA{0x4a, 0x1c, 0x1c, 0xff}
	colEN      = color.RGBA{0x3f, 0xa7, 0xd6, 0xff}
	colENBack  = color.RGBA{0x15, 0x36, 0x47, 0xff}
	colShade   = color.RGBA{0, 0, 0, 0x8c}
)

// ---------- police ----------
var fontReg, fontBold *text.GoTextFaceSource
var faceCache = map[[2]int]*text.GoTextFace{}

func init() {
	var err error
	if fontReg, err = text.NewGoTextFaceSource(bytes.NewReader(goregular.TTF)); err != nil {
		log.Fatal(err)
	}
	if fontBold, err = text.NewGoTextFaceSource(bytes.NewReader(gobold.TTF)); err != nil {
		log.Fatal(err)
	}
}

func face(size int, bold bool) *text.GoTextFace {
	b := 0
	if bold {
		b = 1
	}
	k := [2]int{size, b}
	if f, ok := faceCache[k]; ok {
		return f
	}
	src := fontReg
	if bold {
		src = fontBold
	}
	f := &text.GoTextFace{Source: src, Size: float64(size)}
	faceCache[k] = f
	return f
}

const (
	alignLeft = iota
	alignCenter
	alignRight
)

// txt écrit une ligne de texte ; (x, y) = coin haut de la ligne.
func txt(dst *ebiten.Image, s string, x, y float64, size int, bold bool, c color.Color, align int) {
	op := &text.DrawOptions{}
	op.GeoM.Translate(x, y)
	op.ColorScale.ScaleWithColor(c)
	switch align {
	case alignCenter:
		op.PrimaryAlign = text.AlignCenter
	case alignRight:
		op.PrimaryAlign = text.AlignEnd
	}
	text.Draw(dst, s, face(size, bold), op)
}

// wrap coupe un texte en lignes qui tiennent dans maxW pixels.
func wrap(s string, size int, maxW float64) []string {
	var lines []string
	cur := ""
	for _, w := range strings.Fields(s) {
		try := w
		if cur != "" {
			try = cur + " " + w
		}
		if text.Advance(try, face(size, false)) > maxW && cur != "" {
			lines = append(lines, cur)
			cur = w
		} else {
			cur = try
		}
	}
	if cur != "" {
		lines = append(lines, cur)
	}
	return lines
}

// ---------- formes ----------
func fillRect(dst *ebiten.Image, x, y, w, h float64, c color.Color) {
	vector.DrawFilledRect(dst, float32(x), float32(y), float32(w), float32(h), c, false)
}

func strokeRect(dst *ebiten.Image, x, y, w, h, t float64, c color.Color) {
	vector.StrokeRect(dst, float32(x), float32(y), float32(w), float32(h), float32(t), c, false)
}

func panel(dst *ebiten.Image, x, y, w, h float64) {
	fillRect(dst, x, y, w, h, colPanel)
	strokeRect(dst, x, y, w, h, 2, colBorder)
}

// bar dessine une jauge remplie selon val/max.
func bar(dst *ebiten.Image, x, y, w, h float64, val, max int, fg, bg color.Color) {
	fillRect(dst, x, y, w, h, bg)
	if max > 0 && val > 0 {
		r := float64(val) / float64(max)
		if r > 1 {
			r = 1
		}
		fillRect(dst, x, y, w*r, h, fg)
		// reflet sur le haut de la jauge
		fillRect(dst, x, y, w*r, 2, color.RGBA{0xff, 0xff, 0xff, 0x40})
	}
	strokeRect(dst, x, y, w, h, 1, color.RGBA{0, 0, 0, 0xc0})
}

// ---------------------------------------------------------------
// HUD — barres de vie et d'énergie, toujours visibles en jeu
// ---------------------------------------------------------------
func drawHUD(dst *ebiten.Image, h *Character, showHint bool) {
	if h == nil {
		return
	}
	panel(dst, 8, 8, 236, 102)
	txt(dst, h.Nom, 18, 12, 14, true, colText, alignLeft)
	txt(dst, "Niv. "+strconv.Itoa(h.Niveau), 234, 14, 11, true, colGold, alignRight)
	txt(dst, h.Classe, 18, 31, 11, false, colMuted, alignLeft)

	txt(dst, "PV", 18, 50, 11, true, colHP, alignLeft)
	bar(dst, 46, 52, 122, 11, h.PV, h.PVMax, colHP, colHPBack)
	txt(dst, strconv.Itoa(h.PV)+"/"+strconv.Itoa(h.PVMax), 234, 49, 11, false, colText, alignRight)

	txt(dst, "EN", 18, 69, 11, true, colEN, alignLeft)
	bar(dst, 46, 71, 122, 11, h.Energie, h.EnergieMax, colEN, colENBack)
	txt(dst, strconv.Itoa(h.Energie)+"/"+strconv.Itoa(h.EnergieMax), 234, 68, 11, false, colText, alignRight)

	txt(dst, strconv.Itoa(h.Pieces)+" pièces", 18, 87, 11, true, colGold, alignLeft)
	txt(dst, "Équip. "+strconv.Itoa(h.Equip.Count())+"/3", 234, 87, 11, false, colMuted, alignRight)

	if showHint {
		fillRect(dst, ScreenW-126, ScreenH-28, 118, 20, colShade)
		txt(dst, "E : inventaire", ScreenW-67, ScreenH-26, 12, false, colText, alignCenter)
	}
}

// ---------------------------------------------------------------
// INVENTAIRE
// ---------------------------------------------------------------

// invLayout renvoie le nombre de colonnes et la taille d'un emplacement.
// 5 colonnes pour 10 places ; 10 colonnes une fois l'inventaire agrandi.
func invLayout(capacity int) (cols int, slot, gap float64) {
	if capacity <= 10 {
		return 5, 60, 8
	}
	return 10, 50, 6
}

func drawInventory(dst *ebiten.Image, h *Character, sel int) {
	fillRect(dst, 0, 0, ScreenW, ScreenH, colShade)

	cols, slot, gap := invLayout(h.Capacite)
	rows := (h.Capacite + cols - 1) / cols
	gridW := float64(cols)*slot + float64(cols-1)*gap
	gridH := float64(rows)*slot + float64(rows-1)*gap

	pw := gridW + 48
	if pw < 380 {
		pw = 380
	}
	ph := gridH + 170
	px := (ScreenW - pw) / 2
	py := (ScreenH-ph)/2 + 8 // léger décalage pour ne pas toucher le HUD
	panel(dst, px, py, pw, ph)

	// titre
	txt(dst, "INVENTAIRE", px+22, py+16, 18, true, colGold, alignLeft)
	count := strconv.Itoa(len(h.Inventaire)) + " / " + strconv.Itoa(h.Capacite)
	cc := colText
	if len(h.Inventaire) >= h.Capacite {
		cc = colHP // plein : en rouge
	}
	txt(dst, count, px+pw-22, py+19, 14, true, cc, alignRight)

	// grille d'emplacements
	gx := px + (pw-gridW)/2
	gy := py + 52
	for i := 0; i < h.Capacite; i++ {
		x := gx + float64(i%cols)*(slot+gap)
		y := gy + float64(i/cols)*(slot+gap)
		fillRect(dst, x, y, slot, slot, colSlot)
		if i < len(h.Inventaire) {
			icon := itemIcon(h.Inventaire[i])
			op := &ebiten.DrawImageOptions{}
			s := (slot - 12) / float64(icon.Bounds().Dx())
			op.GeoM.Scale(s, s)
			op.GeoM.Translate(x+6, y+6)
			dst.DrawImage(icon, op)
		}
		if i == sel {
			strokeRect(dst, x-1, y-1, slot+2, slot+2, 3, colGold)
		} else {
			strokeRect(dst, x, y, slot, slot, 1, colBorder)
		}
	}

	// zone d'information sur l'objet sélectionné
	iy := gy + gridH + 16
	fillRect(dst, px+16, iy, pw-32, 62, colPanelLt)
	if sel < len(h.Inventaire) {
		name := h.Inventaire[sel]
		txt(dst, name, px+28, iy+8, 15, true, colText, alignLeft)
		desc := ItemDesc[name]
		if desc == "" {
			desc = "Objet inconnu."
		}
		for i, l := range wrap(desc, 12, pw-56) {
			if i > 1 {
				break
			}
			txt(dst, l, px+28, iy+30+float64(i)*15, 12, false, colMuted, alignLeft)
		}
	} else {
		txt(dst, "Emplacement vide", px+28, iy+22, 13, false, colMuted, alignLeft)
	}

	txt(dst, "Flèches : choisir   ·   Entrée : utiliser   ·   E : fermer",
		px+pw/2, py+ph-26, 11, false, colMuted, alignCenter)
}

// ---------------------------------------------------------------
// CRÉATION DU PERSONNAGE (tâche 11)
// ---------------------------------------------------------------
func drawCreateBackground(dst *ebiten.Image) {
	dst.Fill(color.RGBA{0x2a, 0x26, 0x10, 0xff})
	// quelques bandes façon papier peint, très discrètes
	for x := 0.0; x < ScreenW; x += 22 {
		fillRect(dst, x, 0, 2, ScreenH, color.RGBA{0x33, 0x2e, 0x14, 0xff})
	}
}

func drawNameStep(dst *ebiten.Image, name string, tick int) {
	drawCreateBackground(dst)
	txt(dst, "NOUVELLE PARTIE", ScreenW/2, 110, 30, true, colGold, alignCenter)
	txt(dst, "Comment t'appelles-tu ?", ScreenW/2, 175, 16, false, colText, alignCenter)

	bw, bh := 360.0, 52.0
	bx, by := (ScreenW-bw)/2, 212.0
	fillRect(dst, bx, by, bw, bh, colSlot)
	strokeRect(dst, bx, by, bw, bh, 2, colGold)

	shown := FormatNom(name)
	txt(dst, shown, ScreenW/2, by+13, 22, true, colText, alignCenter)
	// curseur clignotant juste après le texte
	if (tick/30)%2 == 0 {
		w := text.Advance(shown, face(22, true))
		fillRect(dst, ScreenW/2+w/2+3, by+14, 2, 26, colGold)
	}

	txt(dst, "Lettres uniquement — la majuscule est ajoutée automatiquement",
		ScreenW/2, by+bh+14, 12, false, colMuted, alignCenter)
	txt(dst, "Entrée : valider   ·   Échap : retour", ScreenW/2, 420, 12, false, colMuted, alignCenter)
}

func drawClassStep(dst *ebiten.Image, g *Game) {
	drawCreateBackground(dst)
	txt(dst, "CHOISIS TA CLASSE", ScreenW/2, 26, 26, true, colGold, alignCenter)
	txt(dst, "Explorateur : "+FormatNom(string(g.nameBuf)), ScreenW/2, 64, 13, false, colMuted, alignCenter)

	cw, ch, gap := 232.0, 268.0, 18.0
	total := 3*cw + 2*gap
	x0 := (ScreenW - total) / 2
	y0 := 118.0

	for i, c := range Classes {
		x := x0 + float64(i)*(cw+gap)
		selected := i == g.classSel

		fillRect(dst, x, y0, cw, ch, colPanel)
		if selected {
			strokeRect(dst, x-1, y0-1, cw+2, ch+2, 3, colGold)
		} else {
			strokeRect(dst, x, y0, cw, ch, 2, colBorder)
		}

		// aperçu du personnage : il marche vers nous si la carte est choisie
		frame := 0
		if selected {
			frame = (g.uiTick / animSpeed) % g.nFrames
		}
		spr := g.sheet.SubImage(image.Rect(frame*g.frameW, 0, (frame+1)*g.frameW, g.frameH)).(*ebiten.Image)
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Scale(2, 2)
		op.GeoM.Translate(x+cw/2-float64(g.frameW), y0+14)
		if !selected {
			op.ColorScale.Scale(0.55, 0.55, 0.55, 1)
		}
		dst.DrawImage(spr, op)

		ty := y0 + 14 + float64(g.frameH)*2 + 8
		nameCol := colText
		if selected {
			nameCol = colGold
		}
		txt(dst, c.Nom, x+cw/2, ty, 16, true, nameCol, alignCenter)
		txt(dst, "(classe "+c.Base+")", x+cw/2, ty+22, 11, false, colMuted, alignCenter)

		// statistiques : les jauges sont relatives au maximum du jeu (120)
		sy := ty + 44
		txt(dst, "PV max", x+16, sy, 11, true, colHP, alignLeft)
		txt(dst, strconv.Itoa(c.PVMax), x+cw-16, sy, 11, true, colText, alignRight)
		bar(dst, x+16, sy+16, cw-32, 7, c.PVMax, 120, colHP, colHPBack)
		txt(dst, "Énergie", x+16, sy+28, 11, true, colEN, alignLeft)
		txt(dst, strconv.Itoa(c.EnergieMax), x+cw-16, sy+28, 11, true, colText, alignRight)
		bar(dst, x+16, sy+44, cw-32, 7, c.EnergieMax, 120, colEN, colENBack)

		for j, l := range wrap(c.Desc, 11, cw-32) {
			txt(dst, l, x+cw/2, sy+60+float64(j)*14, 11, false, colMuted, alignCenter)
		}
	}

	txt(dst, "Tu commences avec la moitié de tes PV max.", ScreenW/2, 424, 12, false, colText, alignCenter)
	txt(dst, "← → : choisir   ·   Entrée : commencer   ·   Échap : retour",
		ScreenW/2, 446, 12, false, colMuted, alignCenter)
}

// ---------------------------------------------------------------
// MESSAGE TEMPORAIRE (ex. "Inventaire plein !")
// ---------------------------------------------------------------
func drawToast(dst *ebiten.Image, msg string) {
	if msg == "" {
		return
	}
	w := text.Advance(msg, face(13, true)) + 32
	x := (ScreenW - w) / 2
	panel(dst, x, ScreenH-74, w, 30)
	txt(dst, msg, ScreenW/2, ScreenH-67, 13, true, colText, alignCenter)
}

// ---------------------------------------------------------------
// ICÔNES D'OBJETS — dessinées en pixel art directement dans le code.
// Chaque lettre = une couleur (voir iconPalette), '.' = transparent.
// Pour ajouter un objet : ajoute un dessin 16x16 dans iconArt.
// ---------------------------------------------------------------
var iconPalette = map[byte]color.RGBA{
	'k': {0x1a, 0x18, 0x10, 0xff}, // contour
	'w': {0xe8, 0xf1, 0xf2, 0xff}, // verre
	'c': {0x9f, 0xd8, 0xe8, 0xff}, // eau
	'b': {0x3a, 0x78, 0xc4, 0xff}, // bouchon
	'l': {0xe9, 0xdc, 0xa8, 0xff}, // étiquette
	'y': {0xc8, 0xb4, 0x3a, 0xff}, // moquette
	'Y': {0x9c, 0x8a, 0x26, 0xff}, // moquette sombre
	'g': {0x8a, 0x8d, 0x90, 0xff}, // métal
	'G': {0x5a, 0x5d, 0x60, 0xff}, // métal sombre
	'r': {0xb0, 0x5a, 0x24, 0xff}, // rouille
	'n': {0xf7, 0xfb, 0xff, 0xff}, // néon
	'N': {0xbf, 0xe6, 0xff, 0xff}, // lueur
}

var iconArt = map[string][]string{
	"Almond Water": {
		"......kkkk......",
		"......kbbk......",
		"......kbbk......",
		".......kk.......",
		"......kwwk......",
		".....kwwwwk.....",
		"....kwwwwwwk....",
		"....kwllllwk....",
		"....kwllllwk....",
		"....kcccccck....",
		"....kccccwck....",
		"....kccccwck....",
		"....kcccccck....",
		"....kcccccck....",
		".....kkkkkk.....",
		"................",
	},
	"Morceau de moquette": {
		"................",
		"..kkk..kkkk.....",
		".kyyykkyyyyk....",
		".kyYyyyyYyyyk...",
		"..kyyyYyyyyyk...",
		"..kyyyyyyYyyyk..",
		".kyYyyyyyyyyyk..",
		".kyyyyYyyyYyyk..",
		"..kyyyyyyyyyk...",
		"..kyyYyyyYyyyk..",
		".kyyyyyyyyyyyk..",
		".kyyyYyyyyYyk...",
		"..kyyyyyyyyyk...",
		"...kkyykkyykk...",
		".....kk..kk.....",
		"................",
	},
	"Tuyau rouillé": {
		"................",
		"................",
		"................",
		"..kk........kk..",
		".kGGk......kGGk.",
		".kgGkkkkkkkkgGk.",
		".kgggggrggggggk.",
		".kggrggggggrggk.",
		".kgggggggrgggGk.",
		".kGGGGGGGGGGGGk.",
		".kgGkkkkkkkkgGk.",
		".kGGk......kGGk.",
		"..kk........kk..",
		"................",
		"................",
		"................",
	},
	"Néon cassé": {
		"..........NN....",
		".........NknN...",
		"........NknkN...",
		".......NknkN....",
		"......NknkN.....",
		".....NknkN......",
		"....NknkN.......",
		"....kknkN.......",
		"...NkknN........",
		"...kGkN.........",
		"..kGGk..........",
		".kGGk...........",
		".kkk............",
		"................",
		"....k...........",
		"..k...k.........",
	},
	"Barre de fer": {
		"............kk..",
		"...........kgGk.",
		"..........kgGk..",
		".........kgGk...",
		"........kgGk....",
		".......kgGk.....",
		"......kgGk......",
		".....kgGk.......",
		"....kgGk........",
		"...kgGk.........",
		"..kgGk..........",
		".kgGk...........",
		".kGk............",
		"..k.............",
		"................",
		"................",
	},
}

var iconCache = map[string]*ebiten.Image{}

func itemIcon(name string) *ebiten.Image {
	if img, ok := iconCache[name]; ok {
		return img
	}
	art, ok := iconArt[name]
	if !ok {
		art = []string{ // objet inconnu : une petite caisse
			"................", "................", "..kkkkkkkkkkkk..",
			"..kyyyyyyyyyyk..", "..kyYYYYYYYYyk..", "..kyyyyyyyyyyk..",
			"..kkkkkkkkkkkk..", "..kyyyykkyyyyk..", "..kyyyykkyyyyk..",
			"..kyyyyyyyyyyk..", "..kyYyyyyyyYyk..", "..kyyyyyyyyyyk..",
			"..kkkkkkkkkkkk..", "................", "................",
			"................",
		}
	}
	rgba := image.NewRGBA(image.Rect(0, 0, 16, 16))
	for y, row := range art {
		if y >= 16 {
			break
		}
		for x := 0; x < len(row) && x < 16; x++ {
			if c, ok := iconPalette[row[x]]; ok {
				rgba.SetRGBA(x, y, c)
			}
		}
	}
	img := ebiten.NewImageFromImage(rgba)
	iconCache[name] = img
	return img
}
