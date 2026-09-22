package main

import (
	"math/rand"
	"strconv"
)

// ---------------------------------------------------------------
// COMBAT AU TOUR PAR TOUR (façon Pokémon) — logique pure. L'écran de
// combat (sprites, barres animées, boîte de texte) est dans
// combat_ui.go ; la saisie clavier est gérée dans main.go
// (updateCombat), qui appelle uniquement les méthodes ci-dessous.
// ---------------------------------------------------------------

const (
	PoingDamageMin = 8
	PoingDamageMax = 14

	ClawEnergyCost = 15
	ClawDamageMin  = 14
	ClawDamageMax  = 20

	FireballEnergyCost = 25
	FireballDamageMin  = 22
	FireballDamageMax  = 30

	EnergyRegenPerTurn          = 12 // tâche : l'énergie remonte à chaque tour de combat
	EnergyRegenOutOfCombatTicks = 40 // et un peu hors combat (voir main.go)

	MonsterAttackDamageMin    = 6
	MonsterAttackDamageMax    = 12
	MonsterAttackPatternEvery = 3 // tâche : dégâts doublés tous les 3 tours

	PoisonDamagePerTurn = 8 // potion de poison jetée sur le monstre (tâche : dégâts dans le temps)
	PoisonDurationTurns = 4

	CombatShakeTicks = 12 // durée de la secousse d'écran quand un coup porte
	CombatFlashTicks = 8  // durée du flash blanc
)

// AttackID identifie l'une des trois attaques du joueur.
type AttackID int

const (
	AttackPunch AttackID = iota
	AttackClaw
	AttackFireball
)

// Attack décrit une attaque : son coût en énergie, sa plage de dégâts,
// et le sort qu'il faut avoir acheté pour la débloquer (vide si aucun).
type Attack struct {
	ID             AttackID
	Nom            string
	EnergyCost     int
	DmgMin, DmgMax int
	RequiresSpell  string
}

// Attacks — le Coup de poing est gratuit, la Griffe électrique coûte de
// l'énergie mais est connue dès le départ, la Boule de feu est
// verrouillée tant que son grimoire n'est pas acheté chez le marchand.
var Attacks = []Attack{
	{AttackPunch, "Coup de poing", 0, PoingDamageMin, PoingDamageMax, ""},
	{AttackClaw, "Griffe électrique", ClawEnergyCost, ClawDamageMin, ClawDamageMax, ""},
	{AttackFireball, SpellGrimoire, FireballEnergyCost, FireballDamageMin, FireballDamageMax, SpellGrimoire},
}

// CombatMenu — écran de menu courant (tâche : tour du joueur avec
// attaquer / inventaire).
type CombatMenu int

const (
	MenuMain CombatMenu = iota
	MenuAttack
	MenuBag
)

// CombatResult — état de fin de combat.
type CombatResult int

const (
	ResultOngoing CombatResult = iota
	ResultWin
	ResultLose
	ResultFled
)

// Combat — une rencontre en cours entre le joueur et un monstre.
type Combat struct {
	Player  *Character
	Monster *Monster

	Round  int // compteur de tours (tâche : boucle de combat avec compteur de tours)
	Result CombatResult
	Menu   CombatMenu

	Messages        []string // file de messages à afficher un par un
	PoisonTurnsLeft int

	ShakeTimer int
	FlashTimer int

	FleeBlocked bool // vrai contre le boss (tâche : fuite impossible contre le boss)
}

// NewCombat démarre un combat. Si le monstre est plus rapide que le
// joueur (mission bonus initiative), il frappe en premier.
func NewCombat(player *Character, monster *Monster) *Combat {
	c := &Combat{Player: player, Monster: monster, Menu: MenuMain, FleeBlocked: monster.IsBoss}
	c.startRound()
	return c
}

func (c *Combat) push(msg string) { c.Messages = append(c.Messages, msg) }

// PopMessage retire et renvoie le message le plus ancien de la file.
func (c *Combat) PopMessage() (string, bool) {
	if len(c.Messages) == 0 {
		return "", false
	}
	m := c.Messages[0]
	c.Messages = c.Messages[1:]
	return m, true
}

// HasMessages indique s'il reste des messages à afficher : tant que
// c'est le cas, l'UI bloque les nouvelles actions du joueur.
func (c *Combat) HasMessages() bool { return len(c.Messages) > 0 }

// Tick fait décroître les minuteries visuelles (secousse, flash) ;
// appelé une fois par image depuis main.go.
func (c *Combat) Tick() {
	if c.ShakeTimer > 0 {
		c.ShakeTimer--
	}
	if c.FlashTimer > 0 {
		c.FlashTimer--
	}
}

func randRange(min, max int) int {
	if max <= min {
		return min
	}
	return min + rand.Intn(max-min+1)
}

func monsterLabel(m *Monster) string {
	if m.IsBoss {
		return "le boss"
	}
	return "le monstre"
}

func monsterLabelCap(m *Monster) string {
	if m.IsBoss {
		return "Le boss"
	}
	return "Le monstre"
}

// startRound incrémente le tour et fait agir le monstre en premier
// s'il est plus rapide que le joueur (initiative).
func (c *Combat) startRound() {
	c.Round++
	if c.Monster.Vitesse > c.Player.Vitesse {
		c.monsterAttack()
	}
}

// Attack — tâche : le joueur choisit une attaque. Renvoie une erreur
// personnalisée si le sort est verrouillé ou l'énergie insuffisante.
func (c *Combat) Attack(id AttackID) error {
	if c.Result != ResultOngoing || c.HasMessages() {
		return nil
	}
	var atk Attack
	found := false
	for _, a := range Attacks {
		if a.ID == id {
			atk, found = a, true
			break
		}
	}
	if !found {
		return ErrItemNotFound
	}
	if c.Player.AttackLocked(atk) {
		return GameError("Ce sort n'est pas encore appris.")
	}
	if c.Player.Energie < atk.EnergyCost {
		return GameError("Pas assez d'énergie.")
	}
	c.Player.Energie -= atk.EnergyCost
	dmg := randRange(atk.DmgMin, atk.DmgMax)
	c.Monster.HP -= dmg
	c.push(c.Player.Nom + " utilise " + atk.Nom + " : -" + strconv.Itoa(dmg) + " PV.")
	c.FlashTimer = CombatFlashTicks
	c.afterPlayerAction()
	return nil
}

// UseBagItem — tâche : tour du joueur, option inventaire. Almond Water
// soigne le joueur, la potion de poison empoisonne le monstre.
func (c *Combat) UseBagItem(i int) error {
	if c.Result != ResultOngoing || c.HasMessages() {
		return nil
	}
	if i < 0 || i >= len(c.Player.Inventaire) {
		return ErrItemNotFound
	}
	switch c.Player.Inventaire[i] {
	case "Almond Water":
		msg := c.Player.UseItem(i)
		c.push(msg)
	case "Potion de poison":
		c.Player.RemoveItem(i)
		c.PoisonTurnsLeft = PoisonDurationTurns
		c.push("Potion de poison jetée sur " + monsterLabel(c.Monster) + " !")
	default:
		return GameError("Cet objet ne s'utilise pas en combat.")
	}
	c.afterPlayerAction()
	return nil
}

// Flee — tâche : fuite, impossible contre le boss.
func (c *Combat) Flee() {
	if c.Result != ResultOngoing || c.HasMessages() {
		return
	}
	if c.FleeBlocked {
		c.push("Impossible de fuir ce combat !")
		return
	}
	c.Result = ResultFled
	c.push("Tu prends la fuite.")
}

// afterPlayerAction vérifie si le monstre est mort, sinon le laisse
// riposter s'il n'a pas déjà joué ce tour (cas où il est plus lent).
func (c *Combat) afterPlayerAction() {
	if c.Monster.HP <= 0 {
		c.Monster.HP = 0
		c.Monster.Alive = false
		c.Result = ResultWin
		c.push(monsterLabelCap(c.Monster) + " est vaincu !")
		return
	}
	if c.Player.Vitesse >= c.Monster.Vitesse {
		c.monsterAttack()
		if c.Result != ResultOngoing {
			return
		}
	}
	c.endRound()
}

// monsterAttack fait attaquer le monstre : dégâts doublés tous les 3
// tours (pattern demandé par l'énoncé), aura ajoutée pour le boss.
func (c *Combat) monsterAttack() {
	dmgMin, dmgMax := MonsterAttackDamageMin, MonsterAttackDamageMax
	if c.Monster.IsBoss {
		dmgMin, dmgMax = BossAttackDamageMin, BossAttackDamageMax
	}
	dmg := randRange(dmgMin, dmgMax)
	patternBonus := c.Round%MonsterAttackPatternEvery == 0
	if patternBonus {
		dmg *= 2
	}
	label := monsterLabelCap(c.Monster) + " attaque"
	if patternBonus {
		label += " (motif renforcé)"
	}
	if c.Monster.IsBoss {
		aura := BossAuraDamage(c.Player.Equip.Count())
		if aura > 0 {
			dmg += aura
			label += " + aura"
		}
	}
	c.Player.PV -= dmg
	if c.Player.PV < 0 {
		c.Player.PV = 0
	}
	c.push(label + " : -" + strconv.Itoa(dmg) + " PV.")
	c.ShakeTimer = CombatShakeTicks
	if c.Player.PV <= 0 {
		c.Result = ResultLose
	}
}

// endRound applique le poison, régénère l'énergie et ouvre le tour
// suivant (tâche : boucle de combat avec compteur de tours).
func (c *Combat) endRound() {
	if c.PoisonTurnsLeft > 0 {
		c.Monster.HP -= PoisonDamagePerTurn
		c.PoisonTurnsLeft--
		c.push("Le poison ronge " + monsterLabel(c.Monster) + " (-" + strconv.Itoa(PoisonDamagePerTurn) + " PV).")
		if c.Monster.HP <= 0 {
			c.Monster.HP = 0
			c.Monster.Alive = false
			c.Result = ResultWin
			c.push(monsterLabelCap(c.Monster) + " succombe au poison !")
			return
		}
	}
	c.Player.Energie += EnergyRegenPerTurn
	if c.Player.Energie > c.Player.EnergieMax {
		c.Player.Energie = c.Player.EnergieMax
	}
	if c.Player.PV <= 0 {
		c.Result = ResultLose
		return
	}
	c.startRound()
}
