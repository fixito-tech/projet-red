package main

// ---------------------------------------------------------------
// CLAVIER DU COMBAT — lancer un combat, choisir dans les menus
// Attaque / Sac / Fuite, appliquer la fin du combat, et le dialogue
// de confirmation avant le boss. Les règles sont dans combat.go, le
// dessin dans combat_ui.go.
// ---------------------------------------------------------------

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

// startCombat lance une rencontre (tâche : le combat se déclenche au
// contact) avec une courte transition animée.
func (g *Game) startCombat(m *Monster) {
	g.combat = NewCombat(g.hero, m)
	g.openCombatMenu(MenuMain)
	g.combatEnterTimer = CombatEnterTicks
	g.state = StateCombat
}

// updateCombat gère la saisie clavier de l'écran de combat : d'abord
// faire défiler les messages, puis clore le combat s'il est fini, sinon
// laisser le joueur choisir dans le menu ouvert.
func (g *Game) updateCombat() error {
	if g.combatEnterTimer > 0 {
		g.combatEnterTimer--
	}
	c := g.combat
	c.Tick()

	if c.HasMessages() {
		if validatePressed() {
			c.PopMessage()
		}
		return nil
	}

	if c.Result != ResultOngoing {
		if validatePressed() {
			g.endCombat()
		}
		return nil
	}

	switch g.combatMenu {
	case MenuMain:
		g.updateCombatMainMenu()
	case MenuAttack:
		g.updateCombatAttackMenu()
	case MenuBag:
		g.updateCombatBagMenu()
	}
	return nil
}

// updateCombatMainMenu : Attaque, Sac ou Fuite.
func (g *Game) updateCombatMainMenu() {
	navigateList(&g.combatSel, len(combatMainOptions))
	if !validatePressed() {
		return
	}
	switch combatMainOptions[g.combatSel] {
	case "Attaque":
		g.openCombatMenu(MenuAttack)
	case "Sac":
		g.openCombatMenu(MenuBag)
	case "Fuite":
		g.combat.Flee()
	}
}

// updateCombatAttackMenu : choisir une attaque ; si elle est refusée
// (sort verrouillé, énergie insuffisante…), l'erreur s'affiche et le
// joueur peut choisir autre chose.
func (g *Game) updateCombatAttackMenu() {
	navigateList(&g.combatSel, len(Attacks))
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		g.openCombatMenu(MenuMain)
	}
	if !validatePressed() {
		return
	}
	if err := g.combat.Attack(Attacks[g.combatSel].ID); err != nil {
		g.toast(err.Error())
		return
	}
	g.openCombatMenu(MenuMain)
}

// updateCombatBagMenu : choisir un objet utilisable en combat (Almond
// Water, potion de poison).
func (g *Game) updateCombatBagMenu() {
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		g.openCombatMenu(MenuMain)
		return
	}
	entries := combatBagEntries(g.hero)
	if len(entries) == 0 {
		return
	}
	navigateList(&g.combatSel, len(entries))
	if !validatePressed() {
		return
	}
	idx := firstIndexOf(g.hero.Inventaire, entries[g.combatSel].name)
	if idx < 0 {
		return
	}
	if err := g.combat.UseBagItem(idx); err != nil {
		g.toast(err.Error())
		return
	}
	g.openCombatMenu(MenuMain)
}

// endCombat applique les conséquences de la fin d'un combat (victoire,
// défaite, fuite) puis revient au jeu.
func (g *Game) endCombat() {
	c := g.combat
	g.combat = nil
	switch c.Result {
	case ResultWin:
		if g.hero.AddXP(xpRewardFor(c.Monster)) > 0 {
			g.toast("Niveau supérieur !")
		}
		if c.Monster.IsBoss {
			g.bossKilled = true
			g.state = StateVictory
			return
		}
		g.dropLoot(c.Monster)
		g.removeMonster(c.Monster)
	case ResultLose:
		g.hero.Die()
		g.teleportToStart()
		g.toast("Tu es mort... Réveil à moitié de tes PV.")
	}
	// Victoire contre un monstre normal, défaite ou fuite : retour à
	// l'exploration, avec un délai avant qu'un autre monstre attaque.
	g.combatGrace = MonsterCombatGraceTicks
	g.state = StatePlay
}

// updateBossConfirm gère le dialogue "Affronter le boss ? Oui / Non".
func (g *Game) updateBossConfirm() error {
	if leftPressed() || rightPressed() {
		g.bossConfirmSel = 1 - g.bossConfirmSel
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		g.leaveBossDialog()
		return nil
	}
	if validatePressed() {
		if g.bossConfirmSel == 0 {
			g.startCombat(g.boss)
		} else {
			g.leaveBossDialog()
		}
	}
	return nil
}

// dropLoot pose au sol, là où le monstre est mort, les objets qu'il lâche.
func (g *Game) dropLoot(m *Monster) {
	for _, item := range RollLoot(g.rng) {
		g.drops = append(g.drops, &Drop{Room: m.Room, X: m.X, Y: m.Y, Item: item})
	}
}

// xpRewardFor renvoie l'XP gagnée en battant ce monstre (plus pour le boss).
func xpRewardFor(m *Monster) int {
	if m.IsBoss {
		return BossXPReward
	}
	return MonsterXPReward
}

// openCombatMenu affiche un menu de combat avec le premier choix sélectionné.
func (g *Game) openCombatMenu(menu CombatMenu) {
	g.combatMenu, g.combatSel = menu, 0
}

// leaveBossDialog referme le dialogue du boss sans combattre ; il ne se
// rouvre qu'après un court délai, le temps de s'éloigner.
func (g *Game) leaveBossDialog() {
	g.bossGrace = bossDialogGrace
	g.state = StatePlay
}
