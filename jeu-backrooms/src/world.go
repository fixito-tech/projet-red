package main

// ---------------------------------------------------------------
// MONDE — préparer une nouvelle partie (monstres, boss, PNJ), puis ce
// qui se passe tout seul pendant l'exploration : rencontres, délais
// de grâce, énergie qui remonte, monstres qui réapparaissent, loot
// ramassé, retour à la salle de départ.
// ---------------------------------------------------------------

import (
	"math/rand"
	"time"
)

// startNewGame remet le joueur dans la scène de départ, puis peuple la
// carte (monstres, boss, PNJ).
func (g *Game) startNewGame() {
	if !g.placeAtStart() {
		// pas de '@' dans la salle de départ : on démarre au centre
		g.px = float64(GridW / 2 * TileSize)
		g.py = float64(GridH / 2 * TileSize)
	}
	g.dir, g.walkFrame, g.walkTick = DirDown, 0, 0
	g.invOpen, g.invSel = false, 0

	g.populateWorld()

	g.roomGrace, g.combatGrace, g.bossGrace = 0, 0, 0
	g.combat = nil
}

// populateWorld fait apparaître les monstres, le boss et les PNJ sur
// toute la carte, et vide le loot d'une éventuelle partie précédente.
func (g *Game) populateWorld() {
	g.rng = rand.New(rand.NewSource(time.Now().UnixNano()))
	rooms := AllRoomNames()
	spawnable := spawnableRooms(rooms)

	g.monsters = nil
	for i := 0; i < MonsterMaxAlive; i++ {
		if m := SpawnMonster(spawnable, "", g.hero.BasePVMax, g.rng); m != nil {
			g.monsters = append(g.monsters, m)
		}
	}
	g.monsterRespawnTimer = MonsterRespawnDelayTicks
	g.drops = nil

	g.boss = nil
	if bossScene, err := LoadScene(bossRoomName); err == nil {
		row, col := GridH/2, GridW/2
		if r, c, ok := randomFreeTile(bossScene, g.rng); ok {
			row, col = r, c
		}
		g.boss = NewBoss(bossRoomName, float64(col*TileSize), float64(row*TileSize), g.hero.BasePVMax)
	}
	g.bossKilled = false

	g.merchant, g.blacksmith = PlaceNPCs(rooms, g.rng)
}

// updateEncounters fait agir les monstres de la salle affichée et
// déclenche le combat au contact, ou le dialogue du boss. Les délais de
// grâce (après un combat, après un changement de salle) neutralisent le
// contact sans empêcher les monstres de bouger.
func (g *Game) updateEncounters() {
	for _, m := range g.monsters {
		contact := m.Update(g.scene, g.px, g.py)
		if contact && g.roomGrace == 0 && g.combatGrace == 0 {
			g.startCombat(m)
			return
		}
	}

	if g.boss != nil && !g.bossKilled && g.bossGrace == 0 && g.boss.Room == g.scene.Name {
		if boxesOverlap(g.boss.X, g.boss.Y, g.px, g.py) {
			g.bossConfirmSel = 1 // "Non" sélectionné par défaut
			g.state = StateBossConfirm
		}
	}
}

// tickGraceTimers fait décroître les délais de grâce (après un combat
// ou un changement de salle) qui empêchent un monstre de redéclencher
// un combat immédiatement.
func (g *Game) tickGraceTimers() {
	if g.roomGrace > 0 {
		g.roomGrace--
	}
	if g.combatGrace > 0 {
		g.combatGrace--
	}
	if g.bossGrace > 0 {
		g.bossGrace--
	}
}

// updateOutOfCombatRegen — tâche « l'énergie remonte... petit à petit
// hors combat ».
func (g *Game) updateOutOfCombatRegen() {
	if g.hero.Energie >= g.hero.EnergieMax {
		return
	}
	g.energyTick++
	if g.energyTick >= EnergyRegenOutOfCombatTicks {
		g.energyTick = 0
		g.hero.Energie++
	}
}

// updateMonsterRespawn — tâche « réapparition progressive » : tant que
// moins de MonsterMaxAlive monstres sont en vie, un nouveau réapparaît
// après un délai, loin du joueur.
func (g *Game) updateMonsterRespawn() {
	if len(g.monsters) >= MonsterMaxAlive {
		return
	}
	if g.monsterRespawnTimer > 0 {
		g.monsterRespawnTimer--
		return
	}
	rooms := spawnableRooms(AllRoomNames())
	if m := SpawnMonster(rooms, g.scene.Name, g.hero.BasePVMax, g.rng); m != nil {
		g.monsters = append(g.monsters, m)
	}
	g.monsterRespawnTimer = MonsterRespawnDelayTicks
}

// updateDropPickup — tâche « le joueur les ramasse en marchant dessus ».
func (g *Game) updateDropPickup() {
	if len(g.drops) == 0 {
		return
	}
	remaining := g.drops[:0]
	for _, d := range g.drops {
		onDrop := d.Room == g.scene.Name && boxesOverlap(d.X, d.Y, g.px, g.py)
		if onDrop && g.pickUp(d.Item) {
			continue
		}
		remaining = append(remaining, d)
	}
	g.drops = remaining
}

// removeMonster retire un monstre vaincu de la liste des vivants.
func (g *Game) removeMonster(dead *Monster) {
	for i, m := range g.monsters {
		if m == dead {
			g.monsters = append(g.monsters[:i], g.monsters[i+1:]...)
			return
		}
	}
}

// teleportToStart renvoie le joueur à la salle de départ (mort, tâche
// résurrection).
func (g *Game) teleportToStart() {
	g.placeAtStart()
	g.roomGrace = MonsterTransitionGrace
}

// placeAtStart charge la salle de départ et pose le joueur sur son '@'.
// Renvoie false si la salle n'a pas de '@' (le joueur n'est pas déplacé).
func (g *Game) placeAtStart() bool {
	if sc, err := LoadScene(startScene); err == nil {
		g.scene = sc
	}
	if g.scene.Start[0] < 0 {
		return false
	}
	g.px = float64(g.scene.Start[1] * TileSize)
	g.py = float64(g.scene.Start[0] * TileSize)
	return true
}

// pickUp met un objet dans l'inventaire et l'annonce. Renvoie false
// (avec le message "Inventaire plein !") s'il n'y a plus de place.
func (g *Game) pickUp(item string) bool {
	if !g.hero.AddItem(item) {
		g.toast("Inventaire plein !")
		return false
	}
	g.toast("Ramassé : " + item)
	return true
}
