package main

import (
	"math"
	"math/rand"
)

// ---------------------------------------------------------------
// MONSTRES — niveau 1 : le "monstre spaghetti". Logique pure
// (positions, IA, loot). L'affichage est dans monster_draw.go.
// ---------------------------------------------------------------

const (
	// Équilibrage assumé, divergent de l'énoncé (voir README, section
	// « Divergences signalées ») : l'énoncé prévoit 6 monstres et une
	// vie de 2x les PV max du joueur, on joue ici à 9 monstres deux
	// fois moins résistants pour un rythme plus nerveux.
	MonsterMaxAlive          = 9     // monstres vivants au maximum sur toute la carte
	MonsterHPMultiplier      = 1     // vie du monstre = 1x les PV max de base du joueur
	MonsterChaseRadius       = 110.0 // rayon de détection du joueur (px)
	MonsterLoseRadius        = 150.0 // au-delà, le monstre abandonne la poursuite (hystérésis)
	MonsterSpeedFactor       = 0.7   // vitesse du monstre = 70% de celle du joueur (on peut encore fuir)
	MonsterWanderSpeedFactor = 0.30
	MonsterWanderChangeTicks = 90  // change de direction d'errance environ toutes les 1.5s
	MonsterCombatGraceTicks  = 90  // délai de grâce après un combat
	MonsterTransitionGrace   = 45  // délai de grâce après un changement de salle
	MonsterRespawnDelayTicks = 480 // ~8s avant qu'un monstre tué soit remplacé
	MonsterSpawnMinDoorDist  = 3   // cases minimum entre un monstre et une porte/le départ
	MonsterVitesse           = 8   // initiative (mission bonus), plus lent qu'un joueur moyen
	BossVitesse              = 12
)

// AIState — comportement courant d'un monstre.
type AIState int

const (
	AIWander AIState = iota
	AIChase
)

// Monster — un monstre présent quelque part sur la carte.
type Monster struct {
	Room        string // nom canonique de la salle, ex. "level1_r2c3"
	X, Y        float64
	Dir         int
	Frame, Tick int

	AI          AIState
	Seen        bool // affiche le "!" une fois le joueur repéré
	wanderTimer int
	wanderDX    float64
	wanderDY    float64

	PV, PVMax int
	Alive     bool
	IsBoss    bool
	Vitesse   int
}

// NewMonster crée un monstre normal dont la vie dépend des PV max de
// base du joueur (hors bonus d'équipement), comme demandé par l'énoncé.
func NewMonster(room string, x, y float64, playerBasePVMax int) *Monster {
	hp := playerBasePVMax * MonsterHPMultiplier
	return &Monster{
		Room: room, X: x, Y: y, Dir: DirDown,
		PV: hp, PVMax: hp, Alive: true, Vitesse: MonsterVitesse,
	}
}

// ---------------------------------------------------------------
// PLACEMENT — cases libres, loin des portes et du point de départ.
// ---------------------------------------------------------------

// isFreeFloor renvoie vrai si (r,c) est une case de sol assez loin des
// portes pour ne pas coincer le joueur dès son arrivée dans la salle.
func isFreeFloor(scene *Scene, r, c int) bool {
	if r <= 0 || r >= GridH-1 || c <= 0 || c >= GridW-1 {
		return false // on évite les bords (souvent des murs/portes)
	}
	return scene.Grid[r][c] == '.' && !nearDoor(scene, r, c)
}

// nearDoor indique s'il y a une porte '+' à MonsterSpawnMinDoorDist cases
// ou moins de (r,c), en ligne, en colonne ou en diagonale.
func nearDoor(scene *Scene, r, c int) bool {
	d := MonsterSpawnMinDoorDist
	for rr := r - d; rr <= r+d; rr++ {
		for cc := c - d; cc <= c+d; cc++ {
			if scene.tileAt(rr, cc) == '+' {
				return true
			}
		}
	}
	return false
}

// randomFreeTile choisit une case de sol au hasard, loin des portes.
// ok=false si la salle n'en propose aucune (impossible avec nos grilles,
// mais on reste prudent si la carte change).
func randomFreeTile(scene *Scene, rng *rand.Rand) (row, col int, ok bool) {
	var candidates [][2]int
	for r := 0; r < GridH; r++ {
		for c := 0; c < GridW; c++ {
			if isFreeFloor(scene, r, c) {
				candidates = append(candidates, [2]int{r, c})
			}
		}
	}
	if len(candidates) == 0 {
		return 0, 0, false
	}
	p := candidates[rng.Intn(len(candidates))]
	return p[0], p[1], true
}

// randomFreeTileAvoiding est comme randomFreeTile mais évite en plus
// une case précise (typiquement le point de départ, pour ne pas faire
// apparaître le marchand exactement sur le joueur).
func randomFreeTileAvoiding(scene *Scene, rng *rand.Rand, avoid [2]int) (row, col int, ok bool) {
	for tries := 0; tries < 20; tries++ {
		r, c, found := randomFreeTile(scene, rng)
		if !found {
			return 0, 0, false
		}
		if r == avoid[0] && c == avoid[1] {
			continue
		}
		return r, c, true
	}
	return randomFreeTile(scene, rng)
}

// spawnableRooms renvoie la liste des salles où un monstre peut
// apparaître : toute la carte sauf la salle de départ et celle du boss.
// Si la carte ne contient ni l'une ni l'autre (carte de remplacement),
// on retombe simplement sur la liste complète : le jeu doit s'adapter à
// n'importe quelle carte.
func spawnableRooms(all []string) []string {
	var out []string
	for _, name := range all {
		if name == startRoomName || name == bossRoomName {
			continue
		}
		out = append(out, name)
	}
	if len(out) == 0 {
		return all
	}
	return out
}

// SpawnMonster tente de faire apparaître un monstre dans une salle prise
// au hasard parmi les salles autorisées. avoidRoom (optionnel, "" pour
// aucun) permet d'éviter la salle où se trouve le joueur.
func SpawnMonster(rooms []string, avoidRoom string, playerBasePVMax int, rng *rand.Rand) *Monster {
	room, x, y, ok := randomSpot(roomsExcept(rooms, avoidRoom), 8, rng)
	if !ok {
		return nil
	}
	return NewMonster(room, x, y, playerBasePVMax)
}

// roomsExcept retire la salle avoid de la liste, sauf si ça la viderait.
func roomsExcept(rooms []string, avoid string) []string {
	if avoid == "" || len(rooms) <= 1 {
		return rooms
	}
	var out []string
	for _, r := range rooms {
		if r != avoid {
			out = append(out, r)
		}
	}
	if len(out) == 0 {
		return rooms
	}
	return out
}

// randomSpot tire une salle au hasard dans la liste, puis une case libre
// dans cette salle, en réessayant jusqu'à tries fois. Renvoie la salle et
// la position en pixels ; ok=false si aucun essai n'a réussi.
func randomSpot(rooms []string, tries int, rng *rand.Rand) (room string, x, y float64, ok bool) {
	for i := 0; i < tries; i++ {
		name := rooms[rng.Intn(len(rooms))]
		scene, err := LoadScene(name)
		if err != nil {
			continue
		}
		if row, col, found := randomFreeTile(scene, rng); found {
			return name, float64(col * TileSize), float64(row * TileSize), true
		}
	}
	return "", 0, 0, false
}

// ---------------------------------------------------------------
// IA — errance au repos, poursuite si le joueur est proche.
// ---------------------------------------------------------------

// distSquared renvoie la distance au carré entre deux points.
func distSquared(x1, y1, x2, y2 float64) float64 {
	dx, dy := x1-x2, y1-y2
	return dx*dx + dy*dy // carré : suffisant pour comparer à un rayon au carré
}

// Update fait avancer l'IA d'un monstre d'une image. scene est la salle
// affichée, playerX/Y la position du joueur (comme Game.px/py). Renvoie
// vrai si le monstre touche le joueur : le combat doit alors commencer.
func (m *Monster) Update(scene *Scene, playerX, playerY float64) bool {
	if !m.Alive || m.Room != scene.Name {
		return false // on ne simule que les monstres vivants de la salle affichée
	}
	m.updateAIState(distSquared(m.X, m.Y, playerX, playerY))

	var dx, dy float64
	var moving bool
	if m.AI == AIChase {
		dx, dy, moving = m.chaseStep(playerX, playerY)
	} else {
		dx, dy, moving = m.wanderStep()
	}
	m.faceTowards(dx, dy)

	if dx != 0 && canMoveInScene(scene, m.X+dx, m.Y) {
		m.X += dx
	}
	if dy != 0 && canMoveInScene(scene, m.X, m.Y+dy) {
		m.Y += dy
	}
	m.animate(moving)

	return boxesOverlap(m.X, m.Y, playerX, playerY)
}

// updateAIState fait passer le monstre en poursuite quand le joueur
// entre dans son rayon de détection, et le remet en errance quand le
// joueur s'est assez éloigné. d2 = distance au joueur, au carré.
func (m *Monster) updateAIState(d2 float64) {
	if m.AI == AIWander && d2 <= MonsterChaseRadius*MonsterChaseRadius {
		m.AI, m.Seen = AIChase, true
	} else if m.AI == AIChase && d2 > MonsterLoseRadius*MonsterLoseRadius {
		m.AI, m.Seen = AIWander, false
	}
}

// chaseStep renvoie le pas vers le joueur, en ligne droite, à 70 % de
// la vitesse du joueur.
func (m *Monster) chaseStep(playerX, playerY float64) (dx, dy float64, moving bool) {
	vx, vy := playerX-m.X, playerY-m.Y
	if vx == 0 && vy == 0 {
		return 0, 0, false
	}
	speed := playerSpeed * MonsterSpeedFactor
	norm := math.Hypot(vx, vy)
	return vx / norm * speed, vy / norm * speed, true
}

// wanderStep renvoie le pas d'errance, en changeant de direction au
// hasard toutes les ~1,5 s.
func (m *Monster) wanderStep() (dx, dy float64, moving bool) {
	m.wanderTimer--
	if m.wanderTimer <= 0 {
		m.wanderTimer = MonsterWanderChangeTicks
		m.pickWanderDir()
	}
	speed := playerSpeed * MonsterWanderSpeedFactor
	dx, dy = m.wanderDX*speed, m.wanderDY*speed
	return dx, dy, dx != 0 || dy != 0
}

// animate fait défiler les 4 images de marche tant que le monstre
// bouge, et le remet en position immobile sinon.
func (m *Monster) animate(moving bool) {
	if !moving {
		m.Frame, m.Tick = 0, 0
		return
	}
	m.Tick++
	if m.Tick >= animSpeed {
		m.Tick = 0
		m.Frame = (m.Frame + 1) % 4
	}
}

// faceTowards tourne le monstre vers la direction où il se déplace.
func (m *Monster) faceTowards(dx, dy float64) {
	if dx == 0 && dy == 0 {
		return
	}
	if math.Abs(dx) > math.Abs(dy) {
		if dx < 0 {
			m.Dir = DirLeft
		} else {
			m.Dir = DirRight
		}
	} else {
		if dy < 0 {
			m.Dir = DirUp
		} else {
			m.Dir = DirDown
		}
	}
}

// pickWanderDir choisit au hasard une direction d'errance, ou l'arrêt.
func (m *Monster) pickWanderDir() {
	dirs := [][2]float64{{0, 0}, {1, 0}, {-1, 0}, {0, 1}, {0, -1}}
	pick := dirs[rand.Intn(len(dirs))]
	m.wanderDX, m.wanderDY = pick[0], pick[1]
}

// ---------------------------------------------------------------
// LOOT — jamais plus de 3 types différents à la fois par monstre.
// ---------------------------------------------------------------

const MonsterLootMaxTypes = 3

// LootEntry — un type d'objet que peut lâcher un monstre, avec sa
// probabilité d'apparition (indépendante des autres entrées).
type LootEntry struct {
	Item   string
	Chance float64
}

var MonsterLootTable = []LootEntry{
	{"Morceau de moquette", 0.6},
	{"Tuyau rouillé", 0.4},
	{"Néon cassé", 0.3},
	{"Barre de fer", 0.35},
}

// RollLoot tire au hasard les objets lâchés par un monstre à sa mort :
// jamais plus de MonsterLootMaxTypes types différents.
func RollLoot(rng *rand.Rand) []string {
	var drops []string
	for _, e := range MonsterLootTable {
		if len(drops) >= MonsterLootMaxTypes {
			break
		}
		if rng.Float64() < e.Chance {
			drops = append(drops, e.Item)
		}
	}
	return drops
}

// Drop — un objet posé au sol, ramassé en marchant dessus.
type Drop struct {
	Room string
	X, Y float64
	Item string
}
