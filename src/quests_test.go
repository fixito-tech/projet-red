package main

import "testing"

func TestQuestProgressAndReward(t *testing.T) {
	c := newTestCharacter()
	gold := c.Gold

	updateQuests(&c, "kill", 1)
	if c.QuestProgress["premier_contact"] != 1 {
		t.Errorf("progression attendue 1, got %d", c.QuestProgress["premier_contact"])
	}
	if c.Gold != gold+10 {
		t.Errorf("jetons attendus %d, got %d", gold+10, c.Gold)
	}
	if c.QuestProgress["chasseur"] != 1 {
		t.Errorf("la quête Chasseur devait être à 1/3, got %d", c.QuestProgress["chasseur"])
	}
}

func TestQuestCompletesOnlyOnce(t *testing.T) {
	c := newTestCharacter()
	updateQuests(&c, "kill", 1)
	gold := c.Gold

	updateQuests(&c, "kill", 1)
	if c.QuestProgress["premier_contact"] != 1 {
		t.Errorf("une quête terminée ne doit plus progresser, got %d", c.QuestProgress["premier_contact"])
	}
	if c.Gold != gold {
		t.Errorf("la récompense ne doit être donnée qu'une fois : %d -> %d", gold, c.Gold)
	}
}

func TestQuestXPReward(t *testing.T) {
	c := newTestCharacter()
	for i := 0; i < 3; i++ {
		updateQuests(&c, "kill", 1)
	}
	if !questDone(&c, questCatalog[1]) {
		t.Fatal("la quête Chasseur aurait dû être terminée")
	}
	if c.Experience != 50 {
		t.Errorf("expérience attendue 50, got %d", c.Experience)
	}
}
