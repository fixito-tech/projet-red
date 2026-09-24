package main

// ---------------------------------------------------------------
// AFFICHAGE DE L'ÉCONOMIE — PNJ (marchand, forgeron), bulle d'aide,
// menus d'achat/vente/artisanat. Aucune règle de jeu ici : tout est
// dans economy.go / equipment.go. La navigation clavier est gérée
// dans economy_input.go (updateShop / updateCraft).
// ---------------------------------------------------------------

import (
	"strconv"

	"github.com/hajimehoshi/ebiten/v2"
)

const (
	npcFrameW = 40
	npcFrameH = 48
)

var merchantSheet, blacksmithSheet *ebiten.Image

// ensureNPCSprites charge les images des PNJ à la première utilisation :
// d'abord le fichier PNG s'il existe, sinon le visuel de secours.
func ensureNPCSprites() {
	if merchantSheet == nil {
		merchantSheet = loadSprite("merchant.png", placeholderNPCSheet(colMerchantCoat, colMerchantAccent))
		blacksmithSheet = loadSprite("blacksmith.png", placeholderNPCSheet(colSmithCoat, colSmithAccent))
	}
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

// drawMerchantMenu dessine le menu du marchand : titre, onglets
// Acheter / Vendre, puis la liste de l'onglet ouvert.
func drawMerchantMenu(dst *ebiten.Image, g *Game) {
	pw, ph := 520.0, 360.0
	px, py := centeredPanel(dst, pw, ph)
	drawNPCMenuTitle(dst, "LE MARCHAND", g.hero.Pieces, px, py, pw)
	drawShopTabs(dst, g.shopMode, px, py, pw)
	drawShopRows(dst, g.merchantRows(), g.npcMenuSel, px, py+78, pw)
	txt(dst, "Flèches : choisir · Entrée : valider · Échap : partir",
		px+pw/2, py+ph-24, 11, false, colMuted, alignCenter)
}

// drawNPCMenuTitle écrit le titre d'un menu de PNJ et les pièces du joueur.
func drawNPCMenuTitle(dst *ebiten.Image, title string, pieces int, px, py, pw float64) {
	txt(dst, title, px+24, py+16, 20, true, colGold, alignLeft)
	txt(dst, strconv.Itoa(pieces)+" pièces", px+pw-24, py+19, 14, true, colGold, alignRight)
}

// drawShopTabs écrit les onglets ACHETER et VENDRE, l'onglet ouvert en doré.
func drawShopTabs(dst *ebiten.Image, mode ShopMode, px, py, pw float64) {
	buyCol, sellCol := colGold, colMuted
	if mode != ShopBuy {
		buyCol, sellCol = colMuted, colGold
	}
	txt(dst, "ACHETER", px+24, py+46, 13, true, buyCol, alignLeft)
	txt(dst, "VENDRE", px+24+110, py+46, 13, true, sellCol, alignLeft)
	txt(dst, "Tab : changer d'onglet", px+pw-24, py+48, 11, false, colMuted, alignRight)
}

// drawShopRows écrit une ligne par article (nom à gauche, prix à droite),
// la ligne sélectionnée surlignée, ou un message si la liste est vide.
func drawShopRows(dst *ebiten.Image, rows []shopRow, sel int, px, gy, pw float64) {
	for i, row := range rows {
		col := colText
		if i == sel {
			fillRect(dst, px+16, gy-2, pw-32, 22, colPanelLt)
			col = colGold
		}
		txt(dst, row.left, px+28, gy, 13, false, col, alignLeft)
		txt(dst, row.right, px+pw-28, gy, 13, false, col, alignRight)
		gy += 24
	}
	if len(rows) == 0 {
		txt(dst, "(rien à vendre pour le moment)", px+28, gy, 12, false, colMuted, alignLeft)
	}
}

// shopRow — une ligne du menu marchand (nom / prix, ou objet / valeur).
type shopRow struct{ left, right string }

// merchantRows construit la liste affichée selon l'onglet courant.
func (g *Game) merchantRows() []shopRow {
	if g.shopMode == ShopBuy {
		return buyRows(g.hero)
	}
	return sellRows(g.hero)
}

// buyRows liste les articles du marchand avec leur prix ("déjà appris"
// pour un grimoire dont le sort est connu).
func buyRows(h *Character) []shopRow {
	var rows []shopRow
	for _, it := range MerchantSells {
		right := strconv.Itoa(it.Prix) + " po"
		if it.IsSpellbook && h.KnowsSpell(it.Nom) {
			right = "déjà appris"
		}
		rows = append(rows, shopRow{it.Nom, right})
	}
	return rows
}

// sellRows liste les objets de l'inventaire que le marchand rachète,
// avec leur prix.
func sellRows(h *Character) []shopRow {
	var rows []shopRow
	for _, i := range h.sellableSlots() {
		name := h.Inventaire[i]
		rows = append(rows, shopRow{name, strconv.Itoa(MerchantBuyPrices[name]) + " po"})
	}
	return rows
}

// ---------------------------------------------------------------
// MENU DU FORGERON — fabrique l'équipement selon les recettes.
// ---------------------------------------------------------------

// drawBlacksmithMenu dessine le menu du forgeron : une carte par recette.
func drawBlacksmithMenu(dst *ebiten.Image, g *Game) {
	pw, ph := 560.0, 380.0
	px, py := centeredPanel(dst, pw, ph)
	drawNPCMenuTitle(dst, "LE FORGERON", g.hero.Pieces, px, py, pw)

	gy := py + 56
	for i, r := range Recipes {
		drawRecipe(dst, g.hero, r, px, gy, pw, i == g.npcMenuSel)
		gy += 82
	}

	txt(dst, "Flèches : choisir · Entrée : fabriquer et équiper · Échap : partir",
		px+pw/2, py+ph-24, 11, false, colMuted, alignCenter)
}

// drawRecipe dessine une recette : la pièce fabriquée et son bonus, les
// matériaux possédés / nécessaires (en rouge s'il en manque), le prix,
// et "porté" si le joueur porte déjà cette pièce.
func drawRecipe(dst *ebiten.Image, h *Character, r Recipe, px, gy, pw float64, selected bool) {
	bg := colPanelLt
	if selected {
		bg = colSlot
	}
	fillRect(dst, px+16, gy, pw-32, 74, bg)
	if selected {
		strokeRect(dst, px+16, gy, pw-32, 74, 2, colGold)
	}

	current := h.Equip.Get(r.Result.Slot)
	worn := current != nil && current.Nom == r.Result.Nom
	nameCol := colText
	if worn {
		nameCol = colMuted
	}
	txt(dst, r.Result.Nom+" ("+r.Result.Slot.String()+")", px+28, gy+8, 14, true, nameCol, alignLeft)
	txt(dst, "+"+strconv.Itoa(r.Result.BonusPV)+" PV max", px+pw-28, gy+9, 12, true, colHP, alignRight)

	drawMaterials(dst, h, r, px+28, gy+30)
	txt(dst, strconv.Itoa(r.Prix)+" po", px+pw-28, gy+52, 12, false, colGold, alignRight)
	if worn {
		txt(dst, "porté", px+pw-28, gy+52-14, 11, false, colMuted, alignRight)
	}
}

// drawMaterials écrit, pour chaque matériau de la recette, "possédés /
// nécessaires" : en rouge s'il en manque, en clair sinon.
func drawMaterials(dst *ebiten.Image, h *Character, r Recipe, x, y float64) {
	for mat, qty := range r.Materials {
		have := countItem(h.Inventaire, mat)
		col := colHP
		if have >= qty {
			col = colText
		}
		txt(dst, mat+" "+ratio(have, qty), x, y, 11, false, col, alignLeft)
		y += 14
	}
}
