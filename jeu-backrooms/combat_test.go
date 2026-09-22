package main

import (
	"strings"
	"testing"
)

// Tâche : trois attaques, une gratuite.
func TestAttackPunchIsFree(t *testing.T) {
	for _, a := range Attacks {
		if a.ID == AttackPunch && a.EnergyCost != 0 {
			t.Errorf("le Coup de poing devrait être gratuit, coût = %d", a.EnergyCost)
		}
	}
}

// Tâche : la Boule de feu est verrouillée tant qu'elle n'a pas été achetée.
func TestFireballLockedUntilLearned(t *testing.T) {
	h := NewCharacter("Test", Classes[0])
	var fireball Attack
	for _, a := range Attacks {
		if a.ID == AttackFireball {
			fireball = a
		}
	}
	if !h.AttackLocked(fireball) {
		t.Error("la Boule de feu devrait être verrouillée avant achat")
	}
	if err := h.LearnSpell(SpellGrimoire); err != nil {
		t.Fatalf("apprentissage refusé : %v", err)
	}
	if h.AttackLocked(fireball) {
		t.Error("la Boule de feu devrait être débloquée après achat")
	}
}

func TestAttackAgainstLockedSpellFails(t *testing.T) {
	h := NewCharacter("Test", Classes[0])
	m := NewMonster("level1_r1c1", 0, 0, h.BasePVMax)
	c := NewCombat(h, m)
	if err := c.Attack(AttackFireball); err == nil {
		t.Error("attaquer avec un sort non appris devrait renvoyer une erreur")
	}
}

// Tâche : boucle de combat avec compteur de tours + motif (dégâts
// doublés tous les 3 tours).
func TestMonsterAttackPatternDoublesEveryThirdTurn(t *testing.T) {
	h := NewCharacter("Test", Classes[0])
	h.Vitesse = 100 // agit toujours en premier : le tour se déroule dans un ordre prévisible
	m := NewMonster("level1_r1c1", 0, 0, h.BasePVMax)
	m.HP = 1_000_000 // ne doit pas mourir pendant le test
	m.Vitesse = 1
	c := NewCombat(h, m)

	sawPattern := false
	for round := 1; round <= 6; round++ {
		if err := c.Attack(AttackPunch); err != nil {
			t.Fatalf("tour %d : attaque refusée : %v", round, err)
		}
		for {
			msg, ok := c.PopMessage()
			if !ok {
				break
			}
			if strings.Contains(msg, "motif renforcé") {
				if round%MonsterAttackPatternEvery != 0 {
					t.Errorf("tour %d : motif renforcé inattendu (pas un multiple de %d)", round, MonsterAttackPatternEvery)
				}
				sawPattern = true
			}
		}
	}
	if !sawPattern {
		t.Error("le motif de dégâts doublés (tous les 3 tours) ne s'est jamais déclenché")
	}
}

// Tâche : potion de poison = dégâts dans le temps sur le monstre.
func TestPoisonPotionDamagesMonsterOverTime(t *testing.T) {
	h := NewCharacter("Test", Classes[0])
	h.Vitesse = 100 // le joueur agit toujours en premier dans ce test
	h.Inventaire = []string{"Potion de poison"}
	m := NewMonster("level1_r1c1", 0, 0, h.BasePVMax)
	m.Vitesse = 1
	c := NewCombat(h, m)
	hpBefore := m.HP
	if err := c.UseBagItem(0); err != nil {
		t.Fatalf("utilisation de la potion refusée : %v", err)
	}
	// Le round se termine dans le même appel (le joueur est plus rapide) :
	// une tique de poison a donc déjà été consommée.
	if c.PoisonTurnsLeft != PoisonDurationTurns-1 {
		t.Errorf("PoisonTurnsLeft = %d, attendu %d", c.PoisonTurnsLeft, PoisonDurationTurns-1)
	}
	for _, ok := c.PopMessage(); ok; _, ok = c.PopMessage() {
	}
	if m.HP >= hpBefore {
		t.Errorf("le monstre devrait avoir perdu des PV au poison (avant %d, après %d)", hpBefore, m.HP)
	}
}

// Tâche : fuite impossible contre le boss. On donne au joueur des PV
// très élevés pour isoler le test de l'éventuelle attaque préventive
// du boss (initiative) : on ne veut tester QUE le blocage de la fuite.
func TestFleeBlockedAgainstBoss(t *testing.T) {
	h := NewCharacter("Test", Classes[0])
	h.PVMax, h.PV = 100000, 100000
	boss := NewBoss(bossRoomName, 0, 0, h.BasePVMax)
	c := NewCombat(h, boss)
	for _, ok := c.PopMessage(); ok; _, ok = c.PopMessage() {
	}
	c.Flee()
	if c.Result == ResultFled {
		t.Error("la fuite ne devrait pas fonctionner contre le boss")
	}
	if msg, ok := c.PopMessage(); !ok || !strings.Contains(msg, "Impossible de fuir") {
		t.Errorf("message attendu sur la fuite bloquée, obtenu %q (ok=%v)", msg, ok)
	}
}

func TestFleeWorksAgainstNormalMonster(t *testing.T) {
	h := NewCharacter("Test", Classes[0])
	h.PVMax, h.PV = 100000, 100000
	h.Vitesse = 100 // agit toujours en premier : pas d'attaque préventive à gérer
	m := NewMonster("level1_r1c1", 0, 0, h.BasePVMax)
	c := NewCombat(h, m)
	c.Flee()
	if c.Result != ResultFled {
		t.Error("la fuite devrait réussir contre un monstre normal")
	}
}

// ---------------------------------------------------------------
// PREUVE : le boss est imbattable sans équipement, et battable une
// fois entièrement équipé. On simule tour par tour à partir des
// CONSTANTES réelles du jeu (pas de tirage aléatoire), en se plaçant
// à chaque fois dans le cas le plus favorable possible au joueur pour
// la version "sans équipement" (a fortiori : si même ce cas perdant
// échoue, aucune stratégie réelle ne peut gagner), et dans le cas le
// plus défavorable pour la version "avec équipement" (si même ce cas
// pessimiste gagne, la vraie partie, avec ses tirages aléatoires,
// gagne aussi).
// ---------------------------------------------------------------

// simulateBossFight renvoie true si le joueur bat le boss avant de
// mourir, avec une stratégie simple : soigner sous la moitié des PV
// max, sinon attaquer avec la meilleure attaque accessible.
//
// bestCaseForPlayer=true modélise le scénario le plus optimiste
// possible (énergie illimitée, sort le plus puissant, dégâts subis
// minimaux) : utilisé pour l'argument "a fortiori" du boss imbattable
// sans équipement (si même ce joueur hypothétique perd, aucune
// stratégie réelle ne peut gagner).
//
// bestCaseForPlayer=false modélise le pire cas de tirages plausible
// pour un joueur qui joue correctement (gère son énergie, soigne au
// bon moment) mais n'a jamais de chance aux dés (dégâts infligés
// minimaux, dégâts subis maximaux, boss qui frappe avant lui à
// chaque tour) : utilisé pour prouver que l'équipement complet suffit
// même dans le pire des cas.
func simulateBossFight(h *Character, potions int, bestCaseForPlayer bool) (playerWins bool, rounds int) {
	boss := NewBoss(bossRoomName, 0, 0, h.BasePVMax)
	aura := BossAuraDamage(h.Equip.Count())

	bossDmg := BossAttackDamageMax
	if bestCaseForPlayer {
		bossDmg = BossAttackDamageMin
	}

	pv := h.PV
	if pv <= 0 {
		pv = h.PVMax
	}
	bossHP := boss.HP
	healThreshold := h.PVMax / 2
	energy := h.EnergieMax

	playerTurn := func() (bossDead bool) {
		if pv < healThreshold && potions > 0 {
			potions--
			pv += SoinAlmondWater
			if pv > h.PVMax {
				pv = h.PVMax
			}
			return false
		}
		var dmg int
		switch {
		case bestCaseForPlayer:
			dmg = FireballDamageMax // scénario hypothétique : énergie illimitée, sort le plus fort
		case energy >= ClawEnergyCost:
			energy -= ClawEnergyCost
			dmg = ClawDamageMin
		default:
			dmg = PoingDamageMin
		}
		bossHP -= dmg
		return bossHP <= 0
	}
	bossTurn := func(round int) (playerDead bool) {
		dmg := bossDmg
		if round%MonsterAttackPatternEvery == 0 {
			dmg *= 2
		}
		dmg += aura
		pv -= dmg
		return pv <= 0
	}

	const maxRounds = 500
	for round := 1; round <= maxRounds; round++ {
		if bestCaseForPlayer {
			// Ordre le plus généreux : le joueur agit avant de subir les
			// dégâts du tour (initiative favorable au joueur).
			if playerTurn() {
				return true, round
			}
			if bossTurn(round) {
				return false, round
			}
		} else {
			// Ordre le plus dur : le boss frappe avant que le joueur
			// n'ait pu agir ce tour-là (initiative défavorable).
			if bossTurn(round) {
				return false, round
			}
			if playerTurn() {
				return true, round
			}
		}
		energy += EnergyRegenPerTurn // tâche : l'énergie remonte à chaque tour de combat
		if energy > h.EnergieMax {
			energy = h.EnergieMax
		}
	}
	return false, maxRounds
}

func TestBossUnbeatableWithoutEquipment(t *testing.T) {
	for _, c := range Classes {
		for _, niveau := range []int{1, 3, 6} {
			h := NewCharacter("Test", c)
			for i := 0; i < niveau-1; i++ {
				h.AddXP(h.XPMax)
			}
			h.PV = h.PVMax // pleine vie au début du combat
			// Inventaire rempli d'Almond Water (capacité maximale, 3 améliorations).
			for i := 0; i < InventaireMaxUpgr; i++ {
				h.UpgradeInventory()
			}
			potions := h.Capacite

			won, rounds := simulateBossFight(h, potions, true) // meilleur cas possible pour le joueur
			if won {
				t.Errorf("%s niveau %d : le boss ne devrait jamais être battable sans équipement (gagné au tour %d)",
					c.Nom, niveau, rounds)
			}
		}
	}
}

func TestBossBeatableWithFullEquipment(t *testing.T) {
	for _, c := range Classes {
		h := NewCharacter("Test", c)
		h.EquipItem(&EquipmentItem{ItemCasqueTuyau.Nom, ItemCasqueTuyau.Slot, ItemCasqueTuyau.BonusPV})
		h.EquipItem(&EquipmentItem{ItemPlastronMoquette.Nom, ItemPlastronMoquette.Slot, ItemPlastronMoquette.BonusPV})
		h.EquipItem(&EquipmentItem{ItemBottesFer.Nom, ItemBottesFer.Slot, ItemBottesFer.BonusPV})
		if h.Equip.Count() != 3 {
			t.Fatalf("%s : équipement non posé (%d/3)", c.Nom, h.Equip.Count())
		}
		if got := BossAuraDamage(h.Equip.Count()); got != 0 {
			t.Fatalf("%s : l'aura devrait être annulée par l'équipement complet, obtenu %d", c.Nom, got)
		}
		h.PV = h.PVMax
		for i := 0; i < InventaireMaxUpgr; i++ {
			h.UpgradeInventory()
		}
		potions := h.Capacite

		won, _ := simulateBossFight(h, potions, false) // pire cas possible pour le joueur
		if !won {
			t.Errorf("%s : le boss devrait être battable avec l'équipement complet, même dans le pire des cas", c.Nom)
		}
	}
}
