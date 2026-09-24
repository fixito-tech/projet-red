package main

import (
	"math/rand"
	"strconv"
)

// ---------------------------------------------------------------
// COMBAT AU TOUR PAR TOUR (façon Pokémon) — logique pure. L'écran de
// combat (sprites, barres animées, boîte de texte) est dans
// combat_ui.go ; la saisie clavier est gérée dans combat_input.go
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

	SouvenirEnergyCost = 20 // Souvenir d'Almond Water : coût en énergie
	SouvenirSoin       = 30 // PV rendus : 37,5 % des 80 PV max de l'Ancien Résident, la classe la plus fragile

	EnergyRegenPerTurn          = 12 // tâche : l'énergie remonte à chaque tour de combat
	EnergyRegenOutOfCombatTicks = 40 // et un peu hors combat (voir world.go)

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
	AttackSouvenir
)

// EffetSort dit ce que fait un sort une fois lancé. Chaque nouveau
// type d'effet ajoute une constante ici et un cas dans applyAttackEffect.
type EffetSort int

const (
	EffetDegats EffetSort = iota // retire des PV au monstre
	EffetSoin                    // rend des PV au joueur, sans dépasser son maximum
)

// Attack décrit une attaque : son coût en énergie, sa plage de dégâts,
// le sort qu'il faut avoir acheté pour la débloquer (vide si aucun),
// et son effet.
type Attack struct {
	ID             AttackID
	Nom            string
	EnergyCost     int
	DmgMin, DmgMax int
	RequiresSpell  string
	Effet          EffetSort
}

// Attacks — le Coup de poing est gratuit, la Griffe électrique coûte de
// l'énergie mais est connue dès le départ, la Boule de feu est
// verrouillée tant que son grimoire n'est pas acheté chez le marchand.
var Attacks = []Attack{
	{AttackPunch, "Coup de poing", 0, PoingDamageMin, PoingDamageMax, "", EffetDegats},
	{AttackClaw, "Griffe électrique", ClawEnergyCost, ClawDamageMin, ClawDamageMax, "", EffetDegats},
	{AttackFireball, SpellGrimoire, FireballEnergyCost, FireballDamageMin, FireballDamageMax, SpellGrimoire, EffetDegats},
	{AttackSouvenir, SpellSouvenir, SouvenirEnergyCost, 0, 0, SpellSouvenir, EffetSoin},
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

	Messages        []string // file de messages à afficher un par un
	PoisonTurnsLeft int

	ShakeTimer int
	FlashTimer int

	FleeBlocked bool // vrai contre le boss (tâche : fuite impossible contre le boss)
}

// NewCombat démarre un combat. Si le monstre est plus rapide que le
// joueur (mission bonus initiative), il frappe en premier.
func NewCombat(player *Character, monster *Monster) *Combat {
	c := &Combat{Player: player, Monster: monster, FleeBlocked: monster.IsBoss}
	c.startRound()
	return c
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
	atk, found := findAttack(id)
	if !found {
		return ErrItemNotFound
	}
	if c.Player.AttackLocked(atk) {
		return GameError("Ce sort n'est pas encore appris.")
	}
	// Même règle que la bouteille d'Almond Water : un soin inutile est
	// refusé avant de payer l'énergie, et le tour n'est pas consommé.
	if atk.Effet == EffetSoin && c.Player.PV >= c.Player.PVMax {
		return GameError("Tes PV sont déjà au maximum.")
	}
	if c.Player.Energie < atk.EnergyCost {
		return GameError("Pas assez d'énergie.")
	}
	c.Player.Energie -= atk.EnergyCost
	c.applyAttackEffect(atk)
	c.afterPlayerAction()
	return nil
}

// applyAttackEffect applique l'effet d'un sort, une fois l'énergie payée.
// Les trois sorts d'origine sont des dégâts : ils passent tous par le
// premier cas, exactement comme avant l'ajout du champ Effet.
func (c *Combat) applyAttackEffect(atk Attack) {
	switch atk.Effet {
	case EffetDegats:
		dmg := randRange(atk.DmgMin, atk.DmgMax)
		c.Monster.PV -= dmg
		c.push(c.Player.Nom + " utilise " + atk.Nom + " : -" + strconv.Itoa(dmg) + " PV.")
		c.FlashTimer = CombatFlashTicks
	case EffetSoin:
		soin := c.Player.heal(SouvenirSoin)
		c.push(c.Player.Nom + " utilise " + atk.Nom + " : +" + strconv.Itoa(soin) + " PV.")
	}
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
	if c.monsterDefeated(" est vaincu !") {
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
	label := monsterLabelCap(c.Monster) + " attaque"
	if c.Round%MonsterAttackPatternEvery == 0 { // motif : dégâts doublés tous les 3 tours
		dmg *= 2
		label += " (motif renforcé)"
	}
	if c.Monster.IsBoss {
		aura := BossAuraDamage(c.Player.Equip.Count())
		if aura > 0 {
			dmg += aura
			label += " + aura"
		}
	}
	c.Player.PV = max(c.Player.PV-dmg, 0)
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
		c.Monster.PV -= PoisonDamagePerTurn
		c.PoisonTurnsLeft--
		c.push("Le poison ronge " + monsterLabel(c.Monster) + " (-" + strconv.Itoa(PoisonDamagePerTurn) + " PV).")
		if c.monsterDefeated(" succombe au poison !") {
			return
		}
	}
	c.Player.Energie = min(c.Player.Energie+EnergyRegenPerTurn, c.Player.EnergieMax)
	if c.Player.PV <= 0 {
		c.Result = ResultLose
		return
	}
	c.startRound()
}

// monsterDefeated termine le combat par une victoire si le monstre n'a
// plus de PV, en affichant "Le monstre" + ending. Renvoie true s'il est mort.
func (c *Combat) monsterDefeated(ending string) bool {
	if c.Monster.PV > 0 {
		return false
	}
	c.Monster.PV = 0
	c.Monster.Alive = false
	c.Result = ResultWin
	c.push(monsterLabelCap(c.Monster) + ending)
	return true
}

// findAttack cherche une attaque par son identifiant dans la liste Attacks.
func findAttack(id AttackID) (Attack, bool) {
	for _, a := range Attacks {
		if a.ID == id {
			return a, true
		}
	}
	return Attack{}, false
}

// push ajoute un message à la file affichée dans la boîte de texte.
func (c *Combat) push(msg string) { c.Messages = append(c.Messages, msg) }

// PopMessage retire le message le plus ancien de la file (celui affiché).
func (c *Combat) PopMessage() {
	if len(c.Messages) > 0 {
		c.Messages = c.Messages[1:]
	}
}

// HasMessages indique s'il reste des messages à afficher : tant que
// c'est le cas, l'UI bloque les nouvelles actions du joueur.
func (c *Combat) HasMessages() bool { return len(c.Messages) > 0 }

// Tick fait décroître les minuteries visuelles (secousse, flash) ;
// appelé une fois par image depuis combat_input.go.
func (c *Combat) Tick() {
	if c.ShakeTimer > 0 {
		c.ShakeTimer--
	}
	if c.FlashTimer > 0 {
		c.FlashTimer--
	}
}

// randRange tire un nombre entier au hasard entre min et max inclus.
func randRange(min, max int) int {
	if max <= min {
		return min
	}
	return min + rand.Intn(max-min+1)
}

// monsterLabel renvoie "le boss" ou "le monstre" (milieu de phrase).
func monsterLabel(m *Monster) string {
	if m.IsBoss {
		return "le boss"
	}
	return "le monstre"
}

// monsterLabelCap renvoie "Le boss" ou "Le monstre" (début de phrase).
func monsterLabelCap(m *Monster) string {
	if m.IsBoss {
		return "Le boss"
	}
	return "Le monstre"
}
