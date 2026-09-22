package main

// ---------------------------------------------------------------
// ÉCRAN DE COMBAT façon Pokémon — décor, sprites, barres de vie,
// boîte de texte, menus. Aucune règle de jeu ici (voir combat.go) :
// ce fichier ne fait que lire l'état de *Combat et le dessiner. La
// saisie clavier est gérée dans main.go (updateCombat).
// ---------------------------------------------------------------

import (
	"image"
	"image/color"
	"strconv"

	"github.com/hajimehoshi/ebiten/v2"
)

const CombatEnterTicks = 24 // durée de la transition animée d'entrée en combat

var (
	colArenaFloor = color.RGBA{0x2c, 0x28, 0x14, 0xff}
	colArenaBack  = color.RGBA{0x18, 0x16, 0x0a, 0xff}
)

// drawCombat dessine l'écran de combat en entier.
func drawCombat(dst *ebiten.Image, g *Game) {
	c := g.combat
	dst.Fill(colArenaBack)
	fillRect(dst, 0, ScreenH*0.55, ScreenW, ScreenH*0.45, colArenaFloor)
	strokeRect(dst, 0, ScreenH*0.55, ScreenW, 2, 2, colBorder)

	shakeX := 0.0
	if c.ShakeTimer > 0 {
		if c.ShakeTimer%4 < 2 {
			shakeX = 6
		} else {
			shakeX = -6
		}
	}

	// monstre en haut à droite
	drawCombatMonster(dst, c.Monster, shakeX)
	drawCombatBar(dst, ScreenW-330, 24, 280, monsterLabelCap(c.Monster), c.Monster.HP, c.Monster.HPMax, true)

	// joueur de dos en bas à gauche
	drawCombatPlayer(dst, g, shakeX)
	drawCombatBar(dst, 50, ScreenH*0.55-64, 280, g.hero.Nom, g.hero.PV, g.hero.PVMax, false)
	drawCombatEnergy(dst, 50, ScreenH*0.55-34, 280, g.hero.Energie, g.hero.EnergieMax)

	// flash blanc quand un coup porte
	if c.FlashTimer > 0 {
		a := uint8(160 * c.FlashTimer / CombatFlashTicks)
		fillRect(dst, 0, 0, ScreenW, ScreenH, color.RGBA{0xff, 0xff, 0xff, a})
	}

	drawCombatTextBox(dst, g)

	// transition d'entrée : fondu depuis le noir
	if g.combatEnterTimer > 0 {
		a := uint8(255 * g.combatEnterTimer / CombatEnterTicks)
		fillRect(dst, 0, 0, ScreenW, ScreenH, color.RGBA{0, 0, 0, a})
	}
}

func drawCombatMonster(dst *ebiten.Image, m *Monster, shakeX float64) {
	ensureMonsterSprites()
	sheet := monsterSheet
	scale := 2.2
	if m.IsBoss {
		sheet = bossSheet
		scale = 2.2 * bossScale / 1.0
	}
	frame := sheet.SubImage(image.Rect(0, 0, monsterFrameW, monsterFrameH)).(*ebiten.Image)
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(-scale, scale) // face le joueur : symétrique du sprite "vu de dos"
	op.GeoM.Translate(ScreenW-140+shakeX, 70)
	dst.DrawImage(frame, op)
}

func drawCombatPlayer(dst *ebiten.Image, g *Game, shakeX float64) {
	// vue de dos : ligne "haut" du spritesheet joueur (DirUp).
	r := image.Rect(0, DirUp*g.frameH, g.frameW, (DirUp+1)*g.frameH)
	frame := g.sheet.SubImage(r).(*ebiten.Image)
	op := &ebiten.DrawImageOptions{}
	scale := 2.6
	op.GeoM.Scale(scale, scale)
	op.GeoM.Translate(140+shakeX, ScreenH*0.55-float64(g.frameH)*scale+20)
	dst.DrawImage(frame, op)
}

func drawCombatBar(dst *ebiten.Image, x, y, w float64, name string, val, max int, right bool) {
	align := alignLeft
	tx := x
	if right {
		align = alignRight
		tx = x + w
	}
	txt(dst, name, tx, y-18, 15, true, colText, align)
	bar(dst, x, y, w, 14, val, max, colHP, colHPBack)
	txt(dst, strconv.Itoa(val)+"/"+strconv.Itoa(max), tx, y+18, 12, false, colText, align)
}

func drawCombatEnergy(dst *ebiten.Image, x, y, w float64, val, max int) {
	txt(dst, "EN", x, y-2, 11, true, colEN, alignLeft)
	bar(dst, x+28, y, w-28, 10, val, max, colEN, colENBack)
}

// ---------------------------------------------------------------
// BOÎTE DE TEXTE + MENUS
// ---------------------------------------------------------------

func drawCombatTextBox(dst *ebiten.Image, g *Game) {
	c := g.combat
	by := ScreenH - 132.0
	panel(dst, 16, by, ScreenW-32, 116)

	if len(c.Messages) > 0 {
		txt(dst, c.Messages[0], 40, by+18, 15, false, colText, alignLeft)
		txt(dst, "Entrée : continuer", ScreenW-40, by+90, 11, false, colMuted, alignRight)
		return
	}

	switch c.Result {
	case ResultWin:
		txt(dst, monsterLabelCap(c.Monster)+" est vaincu !", 40, by+18, 16, true, colGold, alignLeft)
		txt(dst, "Entrée : continuer", ScreenW-40, by+90, 11, false, colMuted, alignRight)
		return
	case ResultLose:
		txt(dst, "Tu perds connaissance...", 40, by+18, 16, true, colHP, alignLeft)
		txt(dst, "Entrée : continuer", ScreenW-40, by+90, 11, false, colMuted, alignRight)
		return
	case ResultFled:
		txt(dst, "Tu as pris la fuite.", 40, by+18, 16, true, colText, alignLeft)
		txt(dst, "Entrée : continuer", ScreenW-40, by+90, 11, false, colMuted, alignRight)
		return
	}

	switch g.combatMenu {
	case MenuMain:
		drawCombatMainMenu(dst, g, by)
	case MenuAttack:
		drawCombatAttackMenu(dst, g, by)
	case MenuBag:
		drawCombatBagMenu(dst, g, by)
	}
}

var combatMainOptions = []string{"Attaque", "Sac", "Fuite"}

func drawCombatMainMenu(dst *ebiten.Image, g *Game, by float64) {
	txt(dst, "Que fait "+g.hero.Nom+" ?", 40, by+18, 15, false, colText, alignLeft)
	for i, opt := range combatMainOptions {
		c := colText
		if i == g.combatSel {
			c = colGold
		}
		txt(dst, opt, ScreenW-260, by+18+float64(i)*28, 15, true, c, alignLeft)
	}
}

func drawCombatAttackMenu(dst *ebiten.Image, g *Game, by float64) {
	for i, a := range Attacks {
		c := colText
		label := a.Nom
		if a.EnergyCost > 0 {
			label += " (" + strconv.Itoa(a.EnergyCost) + " EN)"
		}
		if g.hero.AttackLocked(a) {
			label = "🔒 " + a.Nom + " (verrouillé)"
			c = colMuted
		} else if i == g.combatSel {
			c = colGold
		}
		txt(dst, label, 40, by+16+float64(i)*24, 14, i == g.combatSel, c, alignLeft)
	}
	txt(dst, "Échap : retour", ScreenW-40, by+90, 11, false, colMuted, alignRight)
}

func drawCombatBagMenu(dst *ebiten.Image, g *Game, by float64) {
	names := combatBagEntries(g.hero)
	if len(names) == 0 {
		txt(dst, "Sac vide pour ce combat.", 40, by+18, 14, false, colMuted, alignLeft)
	}
	for i, e := range names {
		c := colText
		if i == g.combatSel {
			c = colGold
		}
		txt(dst, e.name+" x"+strconv.Itoa(e.count), 40, by+16+float64(i)*22, 13, false, c, alignLeft)
	}
	txt(dst, "Échap : retour", ScreenW-40, by+90, 11, false, colMuted, alignRight)
}

type bagEntry struct {
	name  string
	count int
}

// combatBagEntries liste les objets utilisables en combat (Almond
// Water, Potion de poison), regroupés par nom.
func combatBagEntries(h *Character) []bagEntry {
	order := []string{}
	counts := map[string]int{}
	for _, name := range h.Inventaire {
		if name != "Almond Water" && name != "Potion de poison" {
			continue
		}
		if counts[name] == 0 {
			order = append(order, name)
		}
		counts[name]++
	}
	out := make([]bagEntry, len(order))
	for i, n := range order {
		out[i] = bagEntry{n, counts[n]}
	}
	return out
}

// ---------------------------------------------------------------
// DIALOGUE DU BOSS + ÉCRAN DE VICTOIRE
// ---------------------------------------------------------------

func drawBossConfirm(dst *ebiten.Image, g *Game) {
	fillRect(dst, 0, 0, ScreenW, ScreenH, colShade)
	pw, ph := 460.0, 170.0
	px, py := (ScreenW-pw)/2, (ScreenH-ph)/2
	panel(dst, px, py, pw, ph)
	txt(dst, "Une présence immense bloque le passage.", px+pw/2, py+20, 15, true, colGold, alignCenter)
	txt(dst, "Affronter le boss des Backrooms ?", px+pw/2, py+50, 14, false, colText, alignCenter)

	options := []string{"Oui", "Non"}
	for i, opt := range options {
		c := colMuted
		if i == g.bossConfirmSel {
			c = colGold
		}
		txt(dst, opt, px+pw/2+float64(i*120-60), py+100, 16, true, c, alignCenter)
	}
	txt(dst, "← → : choisir · Entrée : valider · Échap : reculer",
		px+pw/2, py+ph-24, 11, false, colMuted, alignCenter)
}

func drawVictory(dst *ebiten.Image, g *Game) {
	dst.Fill(color.RGBA{0x12, 0x10, 0x06, 0xff})
	txt(dst, "TU AS ÉCHAPPÉ AUX BACKROOMS", ScreenW/2, 160, 34, true, colGold, alignCenter)
	txt(dst, "Le boss s'effondre en silence. La moquette cesse de vibrer.",
		ScreenW/2, 220, 15, false, colText, alignCenter)
	txt(dst, g.hero.Nom+" — Niveau "+strconv.Itoa(g.hero.Niveau), ScreenW/2, 260, 14, false, colMuted, alignCenter)
	txt(dst, "Entrée : retour au menu", ScreenW/2, 420, 13, false, colMuted, alignCenter)
}

// firstIndexOf renvoie l'index du premier objet nommé "name" dans
// l'inventaire (utilisé pour traduire une sélection du sac de combat
// en emplacement réel de l'inventaire).
func firstIndexOf(inv []string, name string) int {
	for i, it := range inv {
		if it == name {
			return i
		}
	}
	return -1
}
