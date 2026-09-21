package main

import "fmt"

// Quest décrit une quête proposée au survivant dans les Backrooms.
type Quest struct {
	ID          string
	Name        string
	Description string
	Event       string // "kill", "craft", "drink", "buy"
	Target      int
	RewardGold  int
	RewardXP    int
}

// questCatalog liste les quêtes du jeu, dans leur ordre d'affichage.
var questCatalog = []Quest{
	{"premier_contact", "Premier contact", "Vaincre 1 Hurleur d'entraînement", "kill", 1, 10, 0},
	{"chasseur", "Chasseur de Hurleurs", "Vaincre 3 Hurleurs d'entraînement", "kill", 3, 0, 50},
	{"soif_amande", "Soif d'Amande", "Boire 2 Eaux d'Amande", "drink", 2, 5, 0},
	{"troc", "Le sens du troc", "Faire 3 achats chez le Troqueur", "buy", 3, 0, 25},
	{"bricoleur", "Mains d'or", "Fabriquer 1 équipement chez le Bricoleur", "craft", 1, 0, 35},
}

// questDone indique si la quête est terminée.
func questDone(c *Character, q Quest) bool {
	return c.QuestProgress[q.ID] >= q.Target
}

// updateQuests fait progresser les quêtes liées à l'événement donné.
// Une quête qui atteint son objectif est validée et sa récompense est donnée.
func updateQuests(c *Character, event string, amount int) {
	for _, q := range questCatalog {
		if q.Event != event || questDone(c, q) {
			continue
		}
		if c.QuestProgress == nil {
			c.QuestProgress = map[string]int{}
		}
		c.QuestProgress[q.ID] += amount
		if c.QuestProgress[q.ID] >= q.Target {
			c.QuestProgress[q.ID] = q.Target
			completeQuest(c, q)
		}
	}
}

// completeQuest annonce la quête terminée et donne sa récompense.
func completeQuest(c *Character, q Quest) {
	fmt.Printf("\n*** Quête terminée : %s ***\n", q.Name)
	if q.RewardGold > 0 {
		c.Gold += q.RewardGold
		fmt.Printf("Récompense : +%d jeton(s).\n", q.RewardGold)
	}
	if q.RewardXP > 0 {
		fmt.Printf("Récompense : %d points d'expérience.\n", q.RewardXP)
		gainExperience(c, q.RewardXP)
	}
}

// accessQuests affiche la liste des quêtes et leur progression.
func accessQuests(c *Character) {
	printTitle("QUÊTES")
	for i, q := range questCatalog {
		status := fmt.Sprintf("%d/%d", c.QuestProgress[q.ID], q.Target)
		if questDone(c, q) {
			status = "terminée"
		}
		fmt.Printf("%d. %s [%s]\n", i+1, q.Name, status)
		fmt.Printf("     %s\n", q.Description)
		fmt.Printf("     Récompense : %s\n", rewardText(q))
	}
	printLine()
}

// rewardText décrit la récompense d'une quête.
func rewardText(q Quest) string {
	switch {
	case q.RewardGold > 0 && q.RewardXP > 0:
		return fmt.Sprintf("%d jeton(s) et %d XP", q.RewardGold, q.RewardXP)
	case q.RewardGold > 0:
		return fmt.Sprintf("%d jeton(s)", q.RewardGold)
	default:
		return fmt.Sprintf("%d XP", q.RewardXP)
	}
}
