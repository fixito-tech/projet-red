package main

// ---------------------------------------------------------------
// AFFICHAGE DES MONSTRES — sprites, indicateur "!", objets au sol.
// Ce fichier ne contient aucune règle de jeu (voir monster.go).
// ---------------------------------------------------------------

import (
	"image"

	"github.com/hajimehoshi/ebiten/v2"
)

const (
	monsterFrameW = 56 // nettement plus grand que le joueur (40x48)
	monsterFrameH = 72
	monsterCols   = 4

	bossScale = 1.6 // le boss est encore plus imposant
)

var monsterSheet, bossSheet *ebiten.Image

// ensureMonsterSprites charge les images des monstres à la première
// utilisation : d'abord le fichier PNG s'il existe, sinon le visuel de
// secours dessiné dans le code. Ensuite, elle ne fait plus rien.
func ensureMonsterSprites() {
	if monsterSheet == nil {
		monsterSheet = loadSprite("monster.png", placeholderMonsterSheet(false))
		bossSheet = loadSprite("boss.png", placeholderMonsterSheet(true))
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
