package main

// ---------------------------------------------------------------
// INTERFACE — tout ce qui est dessiné par-dessus le jeu :
// outils de dessin partagés (texte, panneaux, jauges), HUD et
// message temporaire.
// ---------------------------------------------------------------

import (
	"bytes"
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

// init charge les deux polices du jeu (normale et grasse) au démarrage.
func init() {
	var err error
	if fontReg, err = text.NewGoTextFaceSource(bytes.NewReader(goregular.TTF)); err != nil {
		log.Fatal(err)
	}
	if fontBold, err = text.NewGoTextFaceSource(bytes.NewReader(gobold.TTF)); err != nil {
		log.Fatal(err)
	}
}

// face renvoie la police à la taille demandée, gardée en mémoire après le 1er appel.
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
func txt(dst *ebiten.Image, s string, x, y float64, size int, bold bool, col color.Color, align int) {
	op := &text.DrawOptions{}
	op.GeoM.Translate(x, y)
	op.ColorScale.ScaleWithColor(col)
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

// strokeRect dessine le contour d'un rectangle, d'épaisseur t.
func strokeRect(dst *ebiten.Image, x, y, w, h, t float64, c color.Color) {
	vector.StrokeRect(dst, float32(x), float32(y), float32(w), float32(h), float32(t), c, false)
}

// panel dessine un panneau : fond sombre et bordure.
func panel(dst *ebiten.Image, x, y, w, h float64) {
	fillRect(dst, x, y, w, h, colPanel)
	strokeRect(dst, x, y, w, h, 2, colBorder)
}

// centeredPanel assombrit l'écran puis dessine un panneau centré de
// pw x ph pixels. Renvoie le coin haut-gauche du panneau, à partir
// duquel l'appelant place son contenu.
func centeredPanel(dst *ebiten.Image, pw, ph float64) (px, py float64) {
	fillRect(dst, 0, 0, ScreenW, ScreenH, colShade)
	px = (ScreenW - pw) / 2
	py = (ScreenH - ph) / 2
	panel(dst, px, py, pw, ph)
	return px, py
}

// ratio écrit le chiffre qui accompagne une jauge, au format "35/40".
// Attention : le compteur de l'inventaire utilise un format différent,
// avec des espaces ("3 / 10"), et ne passe donc pas par ici.
func ratio(val, max int) string {
	return strconv.Itoa(val) + "/" + strconv.Itoa(max)
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
	txt(dst, ratio(h.PV, h.PVMax), 234, 49, 11, false, colText, alignRight)

	txt(dst, "EN", 18, 69, 11, true, colEN, alignLeft)
	bar(dst, 46, 71, 122, 11, h.Energie, h.EnergieMax, colEN, colENBack)
	txt(dst, ratio(h.Energie, h.EnergieMax), 234, 68, 11, false, colText, alignRight)

	txt(dst, strconv.Itoa(h.Pieces)+" pièces", 18, 87, 11, true, colGold, alignLeft)
	txt(dst, "Équip. "+strconv.Itoa(h.Equip.Count())+"/3", 234, 87, 11, false, colMuted, alignRight)

	if showHint {
		fillRect(dst, ScreenW-126, ScreenH-28, 118, 20, colShade)
		txt(dst, "E : inventaire", ScreenW-67, ScreenH-26, 12, false, colText, alignCenter)
	}
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
