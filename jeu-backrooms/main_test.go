package main

import "testing"

// Test d'intégration : démarrer une partie fait apparaître monstres,
// boss et PNJ sur la vraie carte (chargée depuis assets/scenes), sans
// jamais paniquer. C'est le chemin qu'emprunte le jeu en vrai (bouton
// "Jouer" du menu) ; ebiten.RunGame n'est pas simulable dans un test,
// donc on appelle directement startNewGame sur un *Game.
func TestStartNewGameSetsUpWorld(t *testing.T) {
	g := &Game{hero: NewCharacter("Test", Classes[0])}
	g.startNewGame()

	if g.scene == nil {
		t.Fatal("la scène de départ n'a pas été chargée")
	}
	if g.scene.Name != startRoomName {
		t.Errorf("scène de départ = %q, attendu %q", g.scene.Name, startRoomName)
	}
	if len(g.monsters) == 0 || len(g.monsters) > MonsterMaxAlive {
		t.Errorf("%d monstres apparus, attendu entre 1 et %d", len(g.monsters), MonsterMaxAlive)
	}
	for _, m := range g.monsters {
		if m.Room == startRoomName || m.Room == bossRoomName {
			t.Errorf("un monstre est apparu dans une salle interdite : %s", m.Room)
		}
	}
	if g.boss == nil {
		t.Fatal("le boss n'a pas été placé")
	}
	if g.boss.Room != bossRoomName {
		t.Errorf("salle du boss = %q, attendu %q", g.boss.Room, bossRoomName)
	}
	if g.merchant == nil || g.merchant.Room != startRoomName {
		t.Error("le marchand devrait être placé dans la salle de départ")
	}
	if g.blacksmith == nil {
		t.Error("le forgeron n'a pas été placé")
	}
	if g.blacksmith.Room == startRoomName || g.blacksmith.Room == bossRoomName {
		t.Errorf("le forgeron ne devrait pas être dans la salle de départ ou du boss : %s", g.blacksmith.Room)
	}
}

// Le respawn progressif doit finir par ramener le nombre de monstres
// vivants au maximum après plusieurs morts, sans jamais dépasser.
func TestMonsterRespawnReachesMax(t *testing.T) {
	g := &Game{hero: NewCharacter("Test", Classes[0])}
	g.startNewGame()

	// On tue tous les monstres d'un coup.
	g.monsters = nil
	g.monsterRespawnTimer = 0

	for i := 0; i < MonsterMaxAlive*(MonsterRespawnDelayTicks+1); i++ {
		g.updateMonsterRespawn()
		if len(g.monsters) > MonsterMaxAlive {
			t.Fatalf("trop de monstres après réapparition : %d", len(g.monsters))
		}
	}
	if len(g.monsters) != MonsterMaxAlive {
		t.Errorf("%d monstres après réapparition complète, attendu %d", len(g.monsters), MonsterMaxAlive)
	}
}
