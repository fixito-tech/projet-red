package main

// ---------------------------------------------------------------
// AFFICHAGE DES MONSTRES — sprites, indicateur "!", objets au sol.
// Ce fichier ne contient aucune règle de jeu (voir monster.go).
// ---------------------------------------------------------------

import (
	"image"
	"image/color"
	"math"
	"path/filepath"

	"github.com/hajimehoshi/ebiten/v2"
)

const (
	monsterFrameW = 56 // nettement plus grand que le joueur (40x48)
	monsterFrameH = 72
	monsterCols   = 4

	bossScale = 1.6 // le boss est encore plus imposant
)

// loadSpriteOrPlaceholder charge assets/<file>, ou dessine une image de
// secours si le fichier est absent (même convention que main.go pour
// le personnage : le jeu doit tourner même sans les assets Python).
func loadSpriteOrPlaceholder(file string, placeholder func() *ebiten.Image) *ebiten.Image {
	if img, err := loadPNG(filepath.Join(assetsDir, file)); err == nil {
		return img
	}
	return placeholder()
}

// placeholderMonsterSheet dessine une silhouette de "spaghetti" noire
// faite de filaments emmêlés : 4 images de marche, contour ondulant
// simple (pas besoin de Pillow pour que le jeu tourne).
func placeholderMonsterSheet(dark bool) *ebiten.Image {
	img := image.NewRGBA(image.Rect(0, 0, monsterFrameW*monsterCols, monsterFrameH))
	body := color.RGBA{0x0a, 0x0a, 0x0c, 0xff}
	if dark {
		body = color.RGBA{0x14, 0x02, 0x04, 0xff}
	}
	eye := color.RGBA{0xd6, 0x2b, 0x2b, 0xff}

	for col := 0; col < monsterCols; col++ {
		ox := col * monsterFrameW
		phase := float64(col) * 1.6
		// tête/corps : un blob ovale au centre-haut
		for y := 4; y < 26; y++ {
			half := 14 - abs(float64(y-14))*0.5
			for x := -int(half); x <= int(half); x++ {
				img.Set(ox+monsterFrameW/2+x, y, body)
			}
		}
		// filaments : plusieurs "nouilles" ondulantes qui pendent vers le bas
		for strand := 0; strand < 6; strand++ {
			baseX := monsterFrameW/2 - 18 + strand*7
			for y := 20; y < monsterFrameH-4; y++ {
				wig := int(6 * mathSin(float64(y)*0.25+phase+float64(strand)))
				px := baseX + wig
				for t := 0; t < 3; t++ { // épaisseur du filament
					img.Set(px+t, y, body)
				}
			}
		}
		// deux yeux rouges, seul signe de vie dans la masse noire
		img.Set(ox+monsterFrameW/2-5, 13, eye)
		img.Set(ox+monsterFrameW/2+5, 13, eye)
	}
	return ebiten.NewImageFromImage(img)
}

func abs(f float64) float64 {
	if f < 0 {
		return -f
	}
	return f
}

// mathSin — enveloppe locale pour ne garder qu'un seul point d'import
// de "math" dans ce fichier de dessin.
func mathSin(x float64) float64 { return math.Sin(x) }

var monsterSheet, bossSheet *ebiten.Image

func ensureMonsterSprites() {
	if monsterSheet == nil {
		monsterSheet = loadSpriteOrPlaceholder("monster.png", func() *ebiten.Image { return placeholderMonsterSheet(false) })
	}
	if bossSheet == nil {
		bossSheet = loadSpriteOrPlaceholder("boss.png", func() *ebiten.Image { return placeholderMonsterSheet(true) })
	}
}

// drawMonster affiche un monstre ancré par les pieds, comme le joueur,
// avec un flip horizontal selon la direction (tâche : "sprite retourné
// selon la direction") et un "!" quand il a repéré le joueur.
func drawMonster(dst *ebiten.Image, m *Monster) {
	ensureMonsterSprites()
	sheet := monsterSheet
	scale := 1.0
	if m.IsBoss {
		sheet = bossSheet
		scale = bossScale
	}

	sx := m.Frame * monsterFrameW
	r := image.Rect(sx, 0, sx+monsterFrameW, monsterFrameH)
	frame := sheet.SubImage(r).(*ebiten.Image)

	op := &ebiten.DrawImageOptions{}
	if m.Dir == DirLeft {
		op.GeoM.Scale(-scale, scale)
		op.GeoM.Translate(float64(monsterFrameW)*scale, 0)
	} else {
		op.GeoM.Scale(scale, scale)
	}
	op.GeoM.Translate(
		m.X-(float64(monsterFrameW)*scale-TileSize)/2,
		m.Y-(float64(monsterFrameH)*scale-TileSize),
	)
	dst.DrawImage(frame, op)

	if m.Seen && !m.IsBoss {
		txt(dst, "!", m.X+TileSize/2, m.Y-float64(monsterFrameH)*scale+6, 20, true, colHP, alignCenter)
	}
}

// drawDrop affiche un objet de loot posé au sol (mêmes icônes 16x16
// que l'inventaire).
func drawDrop(dst *ebiten.Image, d *Drop) {
	icon := itemIcon(d.Item)
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(1.5, 1.5)
	op.GeoM.Translate(d.X+4, d.Y+8)
	dst.DrawImage(icon, op)
}
