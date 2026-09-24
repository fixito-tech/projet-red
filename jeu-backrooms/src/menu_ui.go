package main

// ---------------------------------------------------------------
// AFFICHAGE DU MENU TITRE ET DE LA CRÉATION DU PERSONNAGE (tâche 11)
// — saisie du nom, puis choix de la classe. La saisie clavier de ces
// écrans est dans menu_input.go.
// ---------------------------------------------------------------

import (
	"image"
	"image/color"
	"strconv"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
)

// drawMenu dessine l'écran titre : le fond (image ou décor de secours),
// puis les choix du menu, le choix sélectionné étant en doré.
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
		col := color.Color(colMuted)
		if i == g.menuIndex {
			col = colGold
			txt(screen, "›", ScreenW/2-110, y-3, 26, true, colGold, alignCenter)
		}
		txt(screen, item, ScreenW/2, y, 22, true, col, alignCenter)
	}
	txt(screen, "Flèches ou ZQSD  ·  Entrée pour valider  ·  F11 : plein écran",
		ScreenW/2, 440, 12, false, colMuted, alignCenter)
}

// drawCreateBackground peint le fond sombre à rayures des écrans de menu.
func drawCreateBackground(dst *ebiten.Image) {
	dst.Fill(color.RGBA{0x2a, 0x26, 0x10, 0xff})
	// quelques bandes façon papier peint, très discrètes
	for x := 0.0; x < ScreenW; x += 22 {
		fillRect(dst, x, 0, 2, ScreenH, color.RGBA{0x33, 0x2e, 0x14, 0xff})
	}
}

// drawCreate dessine l'étape en cours de la création du personnage :
// d'abord le nom, puis la classe.
func drawCreate(dst *ebiten.Image, g *Game) {
	if g.createStep == 0 {
		drawNameStep(dst, string(g.nameBuf), g.uiTick)
		return
	}
	drawClassStep(dst, g)
}

// drawNameStep dessine la saisie du nom, avec un curseur qui clignote.
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

// Position et taille des trois cartes de classe.
const (
	classCardY   = 118.0
	classCardW   = 232.0
	classCardH   = 268.0
	classCardGap = 18.0
)

// drawClassStep dessine le choix de la classe : trois cartes côte à
// côte, la carte sélectionnée étant encadrée en doré.
func drawClassStep(dst *ebiten.Image, g *Game) {
	drawCreateBackground(dst)
	txt(dst, "CHOISIS TA CLASSE", ScreenW/2, 26, 26, true, colGold, alignCenter)
	txt(dst, "Explorateur : "+FormatNom(string(g.nameBuf)), ScreenW/2, 64, 13, false, colMuted, alignCenter)

	x0 := (ScreenW - (3*classCardW + 2*classCardGap)) / 2
	for i, c := range Classes {
		x := x0 + float64(i)*(classCardW+classCardGap)
		drawClassCard(dst, g, c, x, i == g.classSel)
	}

	txt(dst, "Tu commences avec la moitié de tes PV max.", ScreenW/2, 424, 12, false, colText, alignCenter)
	txt(dst, "← → : choisir   ·   Entrée : commencer   ·   Échap : retour",
		ScreenW/2, 446, 12, false, colMuted, alignCenter)
}

// drawClassCard dessine la carte d'une classe : cadre, aperçu du
// personnage, nom, puis statistiques.
func drawClassCard(dst *ebiten.Image, g *Game, c ClassInfo, x float64, selected bool) {
	fillRect(dst, x, classCardY, classCardW, classCardH, colPanel)
	if selected {
		strokeRect(dst, x-1, classCardY-1, classCardW+2, classCardH+2, 3, colGold)
	} else {
		strokeRect(dst, x, classCardY, classCardW, classCardH, 2, colBorder)
	}
	drawClassPreview(dst, g, x, selected)

	ty := classCardY + 14 + float64(g.frameH)*2 + 8
	nameCol := colText
	if selected {
		nameCol = colGold
	}
	txt(dst, c.Nom, x+classCardW/2, ty, 16, true, nameCol, alignCenter)
	txt(dst, "(classe "+c.Base+")", x+classCardW/2, ty+22, 11, false, colMuted, alignCenter)
	drawClassStats(dst, c, x, ty+44)
}

// drawClassPreview dessine le personnage en grand : il marche vers nous
// si la carte est choisie, il est grisé sinon.
func drawClassPreview(dst *ebiten.Image, g *Game, x float64, selected bool) {
	frame := 0
	if selected {
		frame = (g.uiTick / animSpeed) % playerFrames
	}
	spr := g.sheet.SubImage(image.Rect(frame*g.frameW, 0, (frame+1)*g.frameW, g.frameH)).(*ebiten.Image)
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(2, 2)
	op.GeoM.Translate(x+classCardW/2-float64(g.frameW), classCardY+14)
	if !selected {
		op.ColorScale.Scale(0.55, 0.55, 0.55, 1)
	}
	dst.DrawImage(spr, op)
}

// drawClassStats affiche les PV max et l'énergie de la classe (jauges
// relatives au maximum du jeu, 120), puis sa description.
func drawClassStats(dst *ebiten.Image, c ClassInfo, x, sy float64) {
	txt(dst, "PV max", x+16, sy, 11, true, colHP, alignLeft)
	txt(dst, strconv.Itoa(c.PVMax), x+classCardW-16, sy, 11, true, colText, alignRight)
	bar(dst, x+16, sy+16, classCardW-32, 7, c.PVMax, 120, colHP, colHPBack)
	txt(dst, "Énergie", x+16, sy+28, 11, true, colEN, alignLeft)
	txt(dst, strconv.Itoa(c.EnergieMax), x+classCardW-16, sy+28, 11, true, colText, alignRight)
	bar(dst, x+16, sy+44, classCardW-32, 7, c.EnergieMax, 120, colEN, colENBack)

	for j, l := range wrap(c.Desc, 11, classCardW-32) {
		txt(dst, l, x+classCardW/2, sy+60+float64(j)*14, 11, false, colMuted, alignCenter)
	}
}
