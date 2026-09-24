package main

// ---------------------------------------------------------------
// ÉCRAN D'EXPLORATION — ce que le joueur voit quand il se promène
// dans les salles (décor, PNJ, monstres, joueur, HUD).
// Ce fichier ne fait que dessiner : aucune règle de jeu ici.
// ---------------------------------------------------------------

import (
	"image"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

// drawPlay dessine l'écran d'exploration, du fond vers l'avant : le
// décor, le contenu de la salle, le joueur, puis l'interface par-dessus.
func (g *Game) drawPlay(screen *ebiten.Image) {
	screen.DrawImage(g.scene.Bg, nil)
	g.drawRoomContents(screen)

	// Le sprite peut être plus grand qu'une case : il déborde vers le haut
	// et sur les côtés, mais son "empreinte au sol" reste de 32x32 pixels.
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(
		g.px-float64(g.frameW-TileSize)/2,
		g.py-float64(g.frameH-TileSize),
	)
	screen.DrawImage(g.playerFrame(), op)

	if npc := g.nearbyNPC(); npc != nil {
		drawInteractionBubble(screen, npc, npc.label())
	}

	// Le HUD passe APRÈS l'inventaire pour rester lisible par-dessus le voile :
	// on voit la barre de vie remonter quand on boit une Almond Water.
	if g.invOpen {
		drawInventory(screen, g.hero, g.invSel)
	}
	drawHUD(screen, g.hero, !g.invOpen)

	if g.showDebug {
		g.drawDebugInfo(screen)
	}
}

// drawRoomContents dessine tout ce qui se trouve dans la salle affichée :
// objets au sol, PNJ, monstres et boss.
func (g *Game) drawRoomContents(screen *ebiten.Image) {
	for _, d := range g.drops {
		if d.Room == g.scene.Name {
			drawDrop(screen, d)
		}
	}
	for _, n := range []*NPC{g.merchant, g.blacksmith} {
		if n != nil && n.Room == g.scene.Name {
			drawNPC(screen, n)
		}
	}
	for _, m := range g.monsters {
		if m.Room == g.scene.Name {
			drawMonster(screen, m)
		}
	}
	if g.boss != nil && !g.bossKilled && g.boss.Room == g.scene.Name {
		drawMonster(screen, g.boss)
	}
}

// drawDebugInfo affiche (touche F1) le nom de la salle et ses sorties.
func (g *Game) drawDebugInfo(screen *ebiten.Image) {
	ebitenutil.DebugPrintAt(screen, "scene: "+g.scene.Name, 8, 100)
	y := 116
	for dir, target := range g.scene.Exits {
		ebitenutil.DebugPrintAt(screen, dir+" -> "+target, 8, y)
		y += 14
	}
}

// playerFrame renvoie l'image du sprite correspondant à la direction + l'anim.
func (g *Game) playerFrame() *ebiten.Image {
	sx := g.walkFrame * g.frameW
	sy := g.dir * g.frameH
	r := image.Rect(sx, sy, sx+g.frameW, sy+g.frameH)
	return g.sheet.SubImage(r).(*ebiten.Image)
}
