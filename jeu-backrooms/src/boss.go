package main

// ---------------------------------------------------------------
// BOSS — placé immobile en haut au milieu de la carte (level1_r0c2).
// Sans équipement, il doit être impossible à battre : ce n'est pas un
// simple indicateur "invincible", mais une vraie mécanique de jeu
// (l'aura) qui inflige plus de dégâts par tour qu'une potion n'en
// soigne ; chaque pièce d'équipement en bloque une partie.
// ---------------------------------------------------------------

const (
	BossHPMultiplier = 2 // en plus du x2 des monstres normaux (donc x4 les PV max de base)

	BossAuraDamageBase       = 60 // strictement > SoinAlmondWater (50) : une potion seule ne suffit jamais
	BossAuraReductionPerItem = 20 // 3 pièces d'équipement -> aura ramenée à 0

	BossAttackDamageMin = 14 // dégâts du coup du boss (avant motif x2 tous les 3 tours)
	BossAttackDamageMax = 20
)

// NewBoss crée le boss de fin de niveau. Sa vie est calculée comme un
// monstre normal (2x les PV max de base du joueur) puis multipliée par
// BossHPMultiplier.
func NewBoss(room string, x, y float64, playerBasePVMax int) *Monster {
	m := NewMonster(room, x, y, playerBasePVMax)
	m.PV *= BossHPMultiplier
	m.PVMax = m.PV
	m.IsBoss = true
	m.Vitesse = BossVitesse
	return m
}

// BossAuraDamage renvoie les dégâts d'aura après réduction par les
// pièces d'équipement portées (0 à 3). Ne descend jamais sous 0.
func BossAuraDamage(equippedCount int) int {
	return max(BossAuraDamageBase-equippedCount*BossAuraReductionPerItem, 0)
}
