package main

// ---------------------------------------------------------------
// SORTS / ATTAQUES — tâche « livre de sort appris une seule fois »
// + mission bonus « sorts ». Logique pure.
//
// Trois attaques existent (voir combat.go pour les dégâts et
// l'énergie) : le Coup de poing est toujours disponible, la Griffe
// électrique aussi (deuxième attaque à énergie), et la Boule de feu
// est verrouillée tant que son grimoire n'a pas été acheté chez le
// marchand.
// ---------------------------------------------------------------

// SpellGrimoire est le nom exact du sort vendu par le marchand.
const SpellGrimoire = "Boule de feu"

// KnowsSpell indique si le personnage a déjà appris ce sort.
func (c *Character) KnowsSpell(name string) bool {
	if c.SpellsKnown == nil {
		return false
	}
	return c.SpellsKnown[name]
}

// LearnSpell apprend un sort une bonne fois pour toutes. Renvoie une
// erreur personnalisée si le sort est déjà connu (tâche « sort déjà
// connu »), pour empêcher de le racheter inutilement.
func (c *Character) LearnSpell(name string) error {
	if c.KnowsSpell(name) {
		return ErrSpellAlreadyKnown
	}
	if c.SpellsKnown == nil {
		c.SpellsKnown = make(map[string]bool)
	}
	c.SpellsKnown[name] = true
	return nil
}

// AttackLocked indique si une attaque de la liste Attacks (combat.go)
// est verrouillée pour ce personnage (sort non acheté).
func (c *Character) AttackLocked(a Attack) bool {
	if a.RequiresSpell == "" {
		return false
	}
	return !c.KnowsSpell(a.RequiresSpell)
}
