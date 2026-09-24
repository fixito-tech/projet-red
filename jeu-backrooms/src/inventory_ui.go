package main

// ---------------------------------------------------------------
// AFFICHAGE DE L'INVENTAIRE — le panneau ouvert avec E et les icônes
// des objets (aussi utilisées pour le loot posé au sol). Aucune règle
// de jeu ici : voir inventory.go.
// ---------------------------------------------------------------

import (
	"image"
	"image/color"
	"strconv"

	"github.com/hajimehoshi/ebiten/v2"
)

// invLayout renvoie le nombre de colonnes et la taille d'un emplacement.
// 5 colonnes pour 10 places ; 10 colonnes une fois l'inventaire agrandi.
func invLayout(capacity int) (cols int, slot, gap float64) {
	if capacity <= 10 {
		return 5, 60, 8
	}
	return 10, 50, 6
}

// drawInventory dessine le panneau de l'inventaire : le titre, la grille
// d'emplacements, puis la description de l'objet sélectionné.
func drawInventory(dst *ebiten.Image, h *Character, sel int) {
	fillRect(dst, 0, 0, ScreenW, ScreenH, colShade)

	cols, slot, gap := invLayout(h.Capacite)
	rows := (h.Capacite + cols - 1) / cols
	gridW := float64(cols)*slot + float64(cols-1)*gap
	gridH := float64(rows)*slot + float64(rows-1)*gap

	pw := max(gridW+48, 380)
	ph := gridH + 170
	px := (ScreenW - pw) / 2
	py := (ScreenH-ph)/2 + 8 // léger décalage pour ne pas toucher le HUD
	panel(dst, px, py, pw, ph)
	drawInventoryTitle(dst, h, px, py, pw)

	gx := px + (pw-gridW)/2
	gy := py + 52
	for i := 0; i < h.Capacite; i++ {
		x := gx + float64(i%cols)*(slot+gap)
		y := gy + float64(i/cols)*(slot+gap)
		drawInventorySlot(dst, h, i, x, y, slot, i == sel)
	}

	drawItemInfo(dst, h, sel, px, gy+gridH+16, pw)
	txt(dst, "Flèches : choisir   ·   Entrée : utiliser   ·   E : fermer",
		px+pw/2, py+ph-26, 11, false, colMuted, alignCenter)
}

// drawInventoryTitle écrit "INVENTAIRE" et le compteur d'objets, en
// rouge quand l'inventaire est plein.
func drawInventoryTitle(dst *ebiten.Image, h *Character, px, py, pw float64) {
	txt(dst, "INVENTAIRE", px+22, py+16, 18, true, colGold, alignLeft)
	count := strconv.Itoa(len(h.Inventaire)) + " / " + strconv.Itoa(h.Capacite)
	cc := colText
	if len(h.Inventaire) >= h.Capacite {
		cc = colHP
	}
	txt(dst, count, px+pw-22, py+19, 14, true, cc, alignRight)
}

// drawInventorySlot dessine la case i : le fond, l'icône de l'objet
// s'il y en a un, et un cadre doré si elle est sélectionnée.
func drawInventorySlot(dst *ebiten.Image, h *Character, i int, x, y, slot float64, selected bool) {
	fillRect(dst, x, y, slot, slot, colSlot)
	if i < len(h.Inventaire) {
		icon := itemIcon(h.Inventaire[i])
		op := &ebiten.DrawImageOptions{}
		s := (slot - 12) / float64(icon.Bounds().Dx())
		op.GeoM.Scale(s, s)
		op.GeoM.Translate(x+6, y+6)
		dst.DrawImage(icon, op)
	}
	if selected {
		strokeRect(dst, x-1, y-1, slot+2, slot+2, 3, colGold)
	} else {
		strokeRect(dst, x, y, slot, slot, 1, colBorder)
	}
}

// drawItemInfo affiche, sous la grille, le nom et la description (2
// lignes maximum) de l'objet sélectionné, ou "Emplacement vide".
func drawItemInfo(dst *ebiten.Image, h *Character, sel int, px, iy, pw float64) {
	fillRect(dst, px+16, iy, pw-32, 62, colPanelLt)
	if sel >= len(h.Inventaire) {
		txt(dst, "Emplacement vide", px+28, iy+22, 13, false, colMuted, alignLeft)
		return
	}
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

// unknownItemArt est l'icône d'un objet sans dessin : une petite caisse.
var unknownItemArt = []string{
	"................",
	"................",
	"..kkkkkkkkkkkk..",
	"..kyyyyyyyyyyk..",
	"..kyYYYYYYYYyk..",
	"..kyyyyyyyyyyk..",
	"..kkkkkkkkkkkk..",
	"..kyyyykkyyyyk..",
	"..kyyyykkyyyyk..",
	"..kyyyyyyyyyyk..",
	"..kyYyyyyyyYyk..",
	"..kyyyyyyyyyyk..",
	"..kkkkkkkkkkkk..",
	"................",
	"................",
	"................",
}

// iconCache garde chaque icône déjà créée, pour ne la dessiner qu'une fois.
var iconCache = map[string]*ebiten.Image{}

// itemIcon renvoie l'icône 16x16 d'un objet (une caisse s'il n'a pas de dessin).
func itemIcon(name string) *ebiten.Image {
	if img, ok := iconCache[name]; ok {
		return img
	}
	art, ok := iconArt[name]
	if !ok {
		art = unknownItemArt
	}
	img := ebiten.NewImageFromImage(pixelArt(art))
	iconCache[name] = img
	return img
}

// pixelArt transforme un dessin 16x16 (une lettre = une couleur de
// iconPalette) en image. Une lettre absente de la palette, comme '.',
// donne une couleur vide : le pixel reste transparent.
func pixelArt(art []string) *image.RGBA {
	rgba := image.NewRGBA(image.Rect(0, 0, 16, 16))
	for y := 0; y < len(art) && y < 16; y++ {
		for x := 0; x < len(art[y]) && x < 16; x++ {
			rgba.SetRGBA(x, y, iconPalette[art[y][x]])
		}
	}
	return rgba
}
