package main

import (
	"math/rand"
	"testing"
)

// Tâche : vie du monstre = 2x les PV max du joueur (hors équipement).
func TestMonsterHPFormula(t *testing.T) {
	m := NewMonster("level1_r1c1", 0, 0, 100)
	if m.HP != 200 || m.HPMax != 200 {
		t.Errorf("HP = %d, attendu 200", m.HP)
	}
}

// Le boss est nettement plus fort qu'un monstre normal.
func TestBossHPStrongerThanMonster(t *testing.T) {
	m := NewMonster("level1_r1c1", 0, 0, 100)
	b := NewBoss(bossRoomName, 0, 0, 100)
	if b.HP <= m.HP {
		t.Errorf("HP du boss (%d) devrait dépasser celui d'un monstre normal (%d)", b.HP, m.HP)
	}
}

// Tâche : jamais plus de 3 types d'objets différents dans le loot.
func TestRollLootMaxThreeTypes(t *testing.T) {
	rng := rand.New(rand.NewSource(1))
	for i := 0; i < 200; i++ {
		drops := RollLoot(rng)
		seen := map[string]bool{}
		for _, d := range drops {
			seen[d] = true
		}
		if len(seen) > MonsterLootMaxTypes {
			t.Fatalf("trop de types différents lâchés : %v", drops)
		}
	}
}

// Tâche : apparition loin des portes.
func TestIsFreeFloorRejectsNearDoors(t *testing.T) {
	scene, err := LoadScene("level1_r0c0")
	if err != nil {
		t.Fatalf("scène introuvable : %v", err)
	}
	for r := 0; r < GridH; r++ {
		for c := 0; c < GridW; c++ {
			if scene.Grid[r][c] != '+' {
				continue
			}
			for dr := -1; dr <= 1; dr++ {
				for dc := -1; dc <= 1; dc++ {
					rr, cc := r+dr, c+dc
					if rr < 0 || rr >= GridH || cc < 0 || cc >= GridW {
						continue
					}
					if isFreeFloor(scene, rr, cc) {
						t.Errorf("(%d,%d) est accepté alors qu'il est à côté d'une porte (%d,%d)", rr, cc, r, c)
					}
				}
			}
		}
	}
}

// spawnableRooms doit exclure le départ et le boss.
func TestSpawnableRoomsExcludesStartAndBoss(t *testing.T) {
	rooms := spawnableRooms(AllRoomNames())
	for _, r := range rooms {
		if r == startRoomName {
			t.Error("la salle de départ ne devrait pas accueillir de monstre")
		}
		if r == bossRoomName {
			t.Error("la salle du boss ne devrait pas accueillir de monstre normal")
		}
	}
}

// Un monstre hors de la salle affichée ne doit pas bouger ni déclencher
// de contact.
func TestMonsterUpdateIgnoresOtherRooms(t *testing.T) {
	scene, err := LoadScene("level1_r1c1")
	if err != nil {
		t.Fatalf("scène introuvable : %v", err)
	}
	m := NewMonster("level1_r2c2", 100, 100, 100)
	contact := m.Update(scene, 100, 100, playerSpeed, false)
	if contact {
		t.Error("un monstre hors de la salle affichée ne devrait jamais être en contact")
	}
	if m.X != 100 || m.Y != 100 {
		t.Error("un monstre hors de la salle affichée ne devrait pas bouger")
	}
}

// Le monstre passe en poursuite quand le joueur entre dans son rayon.
func TestMonsterChasesWithinRadius(t *testing.T) {
	scene, err := LoadScene("level1_r1c1")
	if err != nil {
		t.Fatalf("scène introuvable : %v", err)
	}
	m := NewMonster("level1_r1c1", 200, 200, 100)
	m.Update(scene, 200, 200+MonsterChaseRadius+1, playerSpeed, true) // hors du rayon
	if m.AI != AIWander {
		t.Error("le monstre ne devrait pas encore poursuivre")
	}
	m.Update(scene, 200, 200+10, playerSpeed, true) // très proche
	if m.AI != AIChase || !m.Seen {
		t.Error("le monstre devrait poursuivre et afficher l'indicateur \"!\"")
	}
}

// Le monstre normal se déplace moins vite que le joueur (on doit
// pouvoir fuir).
func TestMonsterSlowerThanPlayer(t *testing.T) {
	if MonsterSpeedFactor >= 1 {
		t.Error("le monstre devrait être plus lent que le joueur pour permettre la fuite")
	}
}
