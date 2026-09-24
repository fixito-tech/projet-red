package main

// ---------------------------------------------------------------
// CLAVIER DU MARCHAND ET DU FORGERON — parler à un PNJ, acheter,
// vendre, fabriquer. Les règles (prix, recettes, erreurs) sont dans
// economy.go, le dessin des menus dans economy_ui.go.
// ---------------------------------------------------------------

import (
	"strconv"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

// nearbyNPC renvoie le PNJ à portée d'interaction dans la salle
// courante, ou nil.
func (g *Game) nearbyNPC() *NPC {
	for _, n := range []*NPC{g.merchant, g.blacksmith} {
		if n == nil || n.Room != g.scene.Name {
			continue
		}
		if distSquared(n.X, n.Y, g.px, g.py) <= NPCInteractRadius*NPCInteractRadius {
			return n
		}
	}
	return nil
}

// openNPC ouvre le menu du marchand ou du forgeron.
func (g *Game) openNPC(n *NPC) {
	g.npcMenuSel = 0
	if n.Kind == NPCMerchant {
		g.shopMode = ShopBuy
		g.state = StateShop
	} else {
		g.state = StateCraft
	}
}

// updateShop gère le menu du marchand : Tab change d'onglet (acheter /
// vendre), Entrée achète ou vend la ligne sélectionnée.
func (g *Game) updateShop() error {
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		g.state = StatePlay
		return nil
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyTab) {
		if g.shopMode == ShopBuy {
			g.shopMode = ShopSell
		} else {
			g.shopMode = ShopBuy
		}
		g.npcMenuSel = 0
	}
	rows := g.merchantRows()
	navigateList(&g.npcMenuSel, len(rows))
	if len(rows) > 0 && g.npcMenuSel >= len(rows) {
		g.npcMenuSel = len(rows) - 1
	}
	if !validatePressed() {
		return nil
	}
	if g.shopMode == ShopBuy {
		g.buySelectedItem()
	} else {
		g.sellSelectedItem()
	}
	return nil
}

// buySelectedItem achète l'article sélectionné dans l'onglet "Acheter".
func (g *Game) buySelectedItem() {
	if g.npcMenuSel >= len(MerchantSells) {
		return
	}
	item := MerchantSells[g.npcMenuSel]
	if err := g.hero.Buy(item); err != nil {
		g.toast(err.Error())
		return
	}
	g.toast("Acheté : " + item.Nom)
}

// sellSelectedItem revend l'objet sélectionné dans l'onglet "Vendre".
func (g *Game) sellSelectedItem() {
	idxs := g.hero.sellableSlots()
	if g.npcMenuSel >= len(idxs) {
		return
	}
	prix, err := g.hero.Sell(idxs[g.npcMenuSel])
	if err != nil {
		g.toast(err.Error())
		return
	}
	g.toast("Vendu pour " + strconv.Itoa(prix) + " pièces")
}

// updateCraft gère le menu du forgeron.
func (g *Game) updateCraft() error {
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		g.state = StatePlay
		return nil
	}
	navigateList(&g.npcMenuSel, len(Recipes))
	if validatePressed() {
		r := Recipes[g.npcMenuSel]
		if err := g.hero.Craft(r); err != nil {
			g.toast(err.Error())
		} else {
			g.toast(r.Result.Nom + " fabriqué et équipé !")
		}
	}
	return nil
}
