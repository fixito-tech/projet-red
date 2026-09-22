package main

// ---------------------------------------------------------------
// AFFICHAGE DE L'ÉCONOMIE — PNJ (marchand, forgeron), bulle d'aide,
// menus d'achat/vente/artisanat. Aucune règle de jeu ici : tout est
// dans economy.go / equipment.go. La navigation clavier est gérée
// dans main.go (updateShop / updateCraft).
// ---------------------------------------------------------------

import (
	"image"
	"image/color"
	"strconv"

	"github.com/hajimehoshi/ebiten/v2"
)

const (
	npcFrameW = 40
	npcFrameH = 48
)

var merchantSheet, blacksmithSheet *ebiten.Image

func ensureNPCSprites() {
	if merchantSheet == nil {
		merchantSheet = loadSpriteOrPlaceholder("merchant.png", func() *ebiten.Image {
			return placeholderNPCSheet(color.RGBA{0x6a, 0x4a, 0x2a, 0xff}, color.RGBA{0xd8, 0xc0, 0x7a, 0xff})
		})
	}
	if blacksmithSheet == nil {
		blacksmithSheet = loadSpriteOrPlaceholder("blacksmith.png", func() *ebiten.Image {
			return placeholderNPCSheet(color.RGBA{0x3a, 0x3a, 0x3d, 0xff}, color.RGBA{0xb0, 0x5a, 0x24, 0xff})
		})
	}
}

// placeholderNPCSheet dessine un PNJ minimal (visuel de secours tant
// que make_monsters.py n'a pas été exécuté avec Pillow).
func placeholderNPCSheet(coat, accent color.RGBA) *ebiten.Image {
	img := image.NewRGBA(image.Rect(0, 0, npcFrameW, npcFrameH))
	skin := color.RGBA{0xe3, 0xbd, 0x92, 0xff}
	for y := 4; y < 14; y++ {
		for x := 13; x < 27; x++ {
			img.Set(x, y, skin)
		}
	}
	for y := 14; y < 44; y++ {
		for x := 8; x < 32; x++ {
			img.Set(x, y, coat)
		}
	}
	for y := 18; y < 30; y++ {
		img.Set(19, y, accent)
		img.Set(20, y, accent)
	}
	return ebiten.NewImageFromImage(img)
}

// drawNPC affiche le marchand ou le forgeron, ancré par les pieds.
func drawNPC(dst *ebiten.Image, n *NPC) {
	ensureNPCSprites()
	sheet := merchantSheet
	if n.Kind == NPCBlacksmith {
		sheet = blacksmithSheet
	}
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(n.X-float64(npcFrameW-TileSize)/2, n.Y-float64(npcFrameH-TileSize))
	dst.DrawImage(sheet, op)
}

// drawInteractionBubble affiche "Appuie sur R pour parler à ..."
// au-dessus d'un PNJ quand le joueur est à portée.
func drawInteractionBubble(dst *ebiten.Image, n *NPC, label string) {
	msg := "R : parler " + label
	w := 200.0
	x := n.X + TileSize/2 - w/2
	y := n.Y - float64(npcFrameH) - 22
	fillRect(dst, x, y, w, 20, colShade)
	txt(dst, msg, n.X+TileSize/2, y+3, 12, false, colText, alignCenter)
}

// ---------------------------------------------------------------
// MENU DU MARCHAND — acheter (articles + grimoire) ou vendre (loot).
// ---------------------------------------------------------------

func drawMerchantMenu(dst *ebiten.Image, g *Game) {
	fillRect(dst, 0, 0, ScreenW, ScreenH, colShade)
	pw, ph := 520.0, 360.0
	px, py := (ScreenW-pw)/2, (ScreenH-ph)/2
	panel(dst, px, py, pw, ph)

	txt(dst, "LE MARCHAND", px+24, py+16, 20, true, colGold, alignLeft)
	txt(dst, strconv.Itoa(g.hero.Pieces)+" pièces", px+pw-24, py+19, 14, true, colGold, alignRight)

	tabAchat, tabVente := "ACHETER", "VENDRE"
	tx := px + 24
	for _, tab := range []string{tabAchat, tabVente} {
		c := colMuted
		if (tab == tabAchat) == (g.shopMode == ShopBuy) {
			c = colGold
		}
		txt(dst, tab, tx, py+46, 13, true, c, alignLeft)
		tx += 110
	}
	txt(dst, "Tab : changer d'onglet", px+pw-24, py+48, 11, false, colMuted, alignRight)

	rows := g.merchantRows()
	gy := py + 78
	for i, row := range rows {
		c := colText
		if i == g.shopSel {
			fillRect(dst, px+16, gy-2, pw-32, 22, colPanelLt)
			c = colGold
		}
		txt(dst, row.left, px+28, gy, 13, false, c, alignLeft)
		txt(dst, row.right, px+pw-28, gy, 13, false, c, alignRight)
		gy += 24
	}
	if len(rows) == 0 {
		txt(dst, "(rien à vendre pour le moment)", px+28, gy, 12, false, colMuted, alignLeft)
	}

	txt(dst, "Flèches : choisir · Entrée : valider · Échap : partir",
		px+pw/2, py+ph-24, 11, false, colMuted, alignCenter)
}

// shopRow — une ligne du menu marchand (nom / prix, ou objet / valeur).
type shopRow struct{ left, right string }

// merchantRows construit la liste affichée selon l'onglet courant.
func (g *Game) merchantRows() []shopRow {
	var rows []shopRow
	if g.shopMode == ShopBuy {
		for _, it := range MerchantSells {
			right := strconv.Itoa(it.Prix) + " po"
			if it.IsSpellbook && g.hero.KnowsSpell(it.Nom) {
				right = "déjà appris"
			}
			rows = append(rows, shopRow{it.Nom, right})
		}
		return rows
	}
	for _, name := range g.hero.Inventaire {
		if prix, ok := MerchantBuyPrices[name]; ok {
			rows = append(rows, shopRow{name, strconv.Itoa(prix) + " po"})
		}
	}
	return rows
}

// ---------------------------------------------------------------
// MENU DU FORGERON — fabrique l'équipement selon les recettes.
// ---------------------------------------------------------------

func drawBlacksmithMenu(dst *ebiten.Image, g *Game) {
	fillRect(dst, 0, 0, ScreenW, ScreenH, colShade)
	pw, ph := 560.0, 380.0
	px, py := (ScreenW-pw)/2, (ScreenH-ph)/2
	panel(dst, px, py, pw, ph)

	txt(dst, "LE FORGERON", px+24, py+16, 20, true, colGold, alignLeft)
	txt(dst, strconv.Itoa(g.hero.Pieces)+" pièces", px+pw-24, py+19, 14, true, colGold, alignRight)

	gy := py + 56
	for i, r := range Recipes {
		selected := i == g.shopSel
		bg := colPanelLt
		if selected {
			bg = colSlot
		}
		fillRect(dst, px+16, gy, pw-32, 74, bg)
		if selected {
			strokeRect(dst, px+16, gy, pw-32, 74, 2, colGold)
		}

		worn := g.hero.Equip.Get(r.Result.Slot) != nil && g.hero.Equip.Get(r.Result.Slot).Nom == r.Result.Nom
		nameCol := colText
		if worn {
			nameCol = colMuted
		}
		txt(dst, r.Result.Nom+" ("+r.Result.Slot.String()+")", px+28, gy+8, 14, true, nameCol, alignLeft)
		txt(dst, "+"+strconv.Itoa(r.Result.BonusPV)+" PV max", px+pw-28, gy+9, 12, true, colHP, alignRight)

		matY := gy + 30
		for mat, qty := range r.Materials {
			have := countItem(g.hero.Inventaire, mat)
			c := colHP
			if have >= qty {
				c = colText
			}
			txt(dst, mat+" "+strconv.Itoa(have)+"/"+strconv.Itoa(qty), px+28, matY, 11, false, c, alignLeft)
			matY += 14
		}
		txt(dst, strconv.Itoa(r.Prix)+" po", px+pw-28, gy+52, 12, false, colGold, alignRight)
		if worn {
			txt(dst, "porté", px+pw-28, gy+52-14, 11, false, colMuted, alignRight)
		}
		gy += 82
	}

	txt(dst, "Flèches : choisir · Entrée : fabriquer et équiper · Échap : partir",
		px+pw/2, py+ph-24, 11, false, colMuted, alignCenter)
}
