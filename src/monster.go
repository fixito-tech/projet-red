package main

import "fmt"

// initGoblin initialise un Hurleur d'entraînement (l'entité utilisée pour les combats d'entraînement).
func initGoblin() Monster {
	return Monster{
		Name:       "Hurleur d'entraînement",
		MaxHP:      40,
		CurrentHP:  40,
		Attack:     5,
		Initiative: 4,
		ExpReward:  30,
	}
}

// goblinPattern joue le tour de combat de l'entité : elle inflige 100% de son attaque,
// sauf tous les 3 tours où elle inflige 200% de son attaque.
func goblinPattern(m *Monster, target *Character, turn int) {
	damage := m.Attack
	if turn%3 == 0 {
		damage = m.Attack * 2
	}

	target.CurrentHP -= damage
	if target.CurrentHP < 0 {
		target.CurrentHP = 0
	}

	fmt.Printf("\n%s inflige à %s %d de dégâts.\n", m.Name, target.Name, damage)
	fmt.Printf("%s : PV %d / %d\n", target.Name, target.CurrentHP, target.MaxHP)
}

/*
                                              .  *
                                        *
                                 .              .
                            *
                       ,
                    ,'
                 ,'
              ,'
           ,'
        ,'
     ★'
   ,'
 *

        même dans les Backrooms, le ciel finit par apparaître.
*/
