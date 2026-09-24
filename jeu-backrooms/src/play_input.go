package main

// ---------------------------------------------------------------
// CLAVIER DE L'EXPLORATION — ce qui se passe à chaque image quand le
// joueur se promène : touches de menu, déplacement, portes vers la
// salle suivante, inventaire ouvert, touches de test (F2 à F4).
// Le dessin de cet écran est dans play_ui.go.
// ---------------------------------------------------------------

import (
	"log"
	"strconv"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

// updatePlay fait avancer l'exploration d'une image : minuteries, touches
// de menu, déplacement, portes, ramassage, puis rencontres. L'ordre compte :
// si une touche ouvre un menu, le joueur ne bouge pas cette image-là.
func (g *Game) updatePlay() error {
	g.updateDebugKeys()
	g.tickGraceTimers()
	g.updateOutOfCombatRegen()
	g.updateMonsterRespawn()

	if g.handlePlayKeys() {
		return nil
	}

	g.movePlayer()

	g.checkTransition()
	g.updateDropPickup()
	g.updateEncounters()
	return nil
}

// handlePlayKeys traite les touches qui ouvrent un menu ou changent
// d'état. Renvoie true quand la frame est consommée : le joueur ne se
// déplace pas ce tour-là.
func (g *Game) handlePlayKeys() bool {
	// Inventaire ouvert : il capte les touches et le joueur ne bouge plus.
	if g.invOpen {
		g.updateInventory()
		g.walkFrame, g.walkTick = 0, 0
		return true
	}
	if npc := g.nearbyNPC(); npc != nil && inpututil.IsKeyJustPressed(ebiten.KeyR) {
		g.openNPC(npc)
		return true
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyE) {
		g.invOpen = true
		return true
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		g.state = StateMenu
		return true
	}
	// F1 ne consomme pas la frame : on peut basculer le debug en marchant.
	if inpututil.IsKeyJustPressed(ebiten.KeyF1) {
		g.showDebug = !g.showDebug
	}
	return false
}

// movePlayer lit les touches de direction, déplace le joueur axe par
// axe et fait avancer l'animation de marche.
func (g *Game) movePlayer() {
	dx, dy := g.readMoveKeys()

	// Déplacement axe par axe : on glisse le long des murs au lieu de se bloquer.
	if dx != 0 && g.canMove(g.px+dx, g.py) {
		g.px += dx
	}
	if dy != 0 && g.canMove(g.px, g.py+dy) {
		g.py += dy
	}

	g.animateWalk(dx != 0 || dy != 0)
}

// readMoveKeys renvoie le déplacement demandé au clavier pendant cette
// image, et tourne le joueur vers la direction appuyée. Ici on regarde
// si la touche est MAINTENUE (IsKeyPressed) : garder le doigt appuyé
// fait avancer en continu.
func (g *Game) readMoveKeys() (dx, dy float64) {
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
	return dx, dy
}

// animateWalk fait défiler les images de marche tant que le joueur
// avance, et le remet en position immobile sinon.
func (g *Game) animateWalk(moving bool) {
	if !moving {
		g.walkFrame, g.walkTick = 0, 0
		return
	}
	g.walkTick++
	if g.walkTick >= animSpeed {
		g.walkTick = 0
		g.walkFrame = (g.walkFrame + 1) % playerFrames
	}
}

// canMove teste la boîte de collision aux quatre coins contre les murs '#'.
func (g *Game) canMove(x, y float64) bool {
	return canMoveInScene(g.scene, x, y)
}

// checkTransition : si le joueur est sur une case '+', on change de scène.
func (g *Game) checkTransition() {
	cx := g.px + boxOffX + boxW/2
	cy := g.py + boxOffY + boxH/2
	col := int(cx) / TileSize
	row := int(cy) / TileSize
	if g.scene.tileAt(row, col) != '+' {
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
	g.roomGrace = MonsterTransitionGrace

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

// updateInventory gère l'inventaire ouvert : E ou Échap le ferme, les
// flèches choisissent une case, Entrée utilise l'objet choisi.
func (g *Game) updateInventory() {
	if inpututil.IsKeyJustPressed(ebiten.KeyE) || inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		g.invOpen = false
		return
	}
	g.moveInventoryCursor()
	if !validatePressed() {
		return
	}
	if m := g.hero.UseItem(g.invSel); m != "" {
		g.toast(m)
	}
}

// moveInventoryCursor déplace la sélection dans la grille : gauche/droite
// d'une case, haut/bas d'une ligne entière, sans jamais en sortir.
func (g *Game) moveInventoryCursor() {
	cols, _, _ := invLayout(g.hero.Capacite)
	if rightPressed() {
		g.invSel++
	}
	if leftPressed() {
		g.invSel--
	}
	if downPressed() {
		g.invSel += cols
	}
	if upPressed() {
		g.invSel -= cols
	}
	g.invSel = min(max(g.invSel, 0), g.hero.Capacite-1)
}

// Touches de test pour la démo (à retirer ou garder pour l'oral).
var testItems = []string{"Morceau de moquette", "Tuyau rouillé", "Néon cassé", "Barre de fer", "Almond Water"}

const (
	debugDamage     = 15 // F3 : PV perdus
	debugEnergyLoss = 10 // F3 : énergie perdue
)

// updateDebugKeys gère les touches de test : F2 ramasse un objet, F3 fait
// perdre des PV et de l'énergie, F4 agrandit l'inventaire.
func (g *Game) updateDebugKeys() {
	h := g.hero
	if inpututil.IsKeyJustPressed(ebiten.KeyF2) { // ajoute un objet
		g.pickUp(testItems[g.uiTick%len(testItems)])
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyF3) { // prend des dégâts
		h.PV = max(h.PV-debugDamage, 0)
		h.Energie = max(h.Energie-debugEnergyLoss, 0)
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyF4) { // agrandit l'inventaire
		if h.UpgradeInventory() {
			g.toast("Inventaire agrandi : " + strconv.Itoa(h.Capacite) + " places")
		} else {
			g.toast("Amélioration maximale atteinte")
		}
	}
}
