package main

import "math/rand"

// ---------------------------------------------------------------
// ÉCONOMIE — marchand (achète le loot, vend des objets) et forgeron
// (fabrique l'équipement). Logique pure : aucun affichage ici (voir
// economy_ui.go).
// ---------------------------------------------------------------

// NPCKind distingue le marchand du forgeron.
type NPCKind int

const (
	NPCMerchant NPCKind = iota
	NPCBlacksmith
)

// NPC — un personnage non joueur immobile, avec sa position au sol.
type NPC struct {
	Kind NPCKind
	Room string
	X, Y float64
}

// label renvoie la fin de la bulle "R : parler au marchand" (ou au forgeron).
func (n *NPC) label() string {
	if n.Kind == NPCMerchant {
		return "au marchand"
	}
	return "au forgeron"
}

// ShopMode — onglet courant du menu marchand (acheter/vendre).
type ShopMode int

const (
	ShopBuy ShopMode = iota
	ShopSell
)

// PlaceNPCs place le marchand au point de départ et le forgeron dans
// une salle aléatoire (ni le départ, ni le boss). Le jeu s'adapte à
// n'importe quelle carte : si la salle de départ est introuvable, le
// marchand n'est simplement pas placé.
func PlaceNPCs(rooms []string, rng *rand.Rand) (merchant, blacksmith *NPC) {
	if scene, err := LoadScene(startRoomName); err == nil {
		if r, c, ok := randomFreeTileAvoiding(scene, rng, scene.Start); ok {
			merchant = &NPC{NPCMerchant, startRoomName, float64(c * TileSize), float64(r * TileSize)}
		}
	}
	if room, x, y, ok := randomSpot(spawnableRooms(rooms), 10, rng); ok {
		blacksmith = &NPC{NPCBlacksmith, room, x, y}
	}
	return
}

// ShopItem — un article vendu par le marchand.
type ShopItem struct {
	Nom         string
	Prix        int
	IsSpellbook bool // vrai pour les grimoires : le sort est appris à l'achat, rien n'entre dans l'inventaire
}

// MerchantSells — articles disponibles chez le marchand (tâche : vend
// les articles de l'énoncé, dont la potion de poison).
var MerchantSells = []ShopItem{
	{"Almond Water", 10, false},
	{"Potion de poison", 20, false},
	{SpellGrimoire, 120, true},
	{SpellSouvenir, 60, true},
}

// MerchantBuyPrices — prix auquel le marchand rachète le loot ramassé
// sur les monstres (jamais plus de 3 types différents, voir monster.go).
var MerchantBuyPrices = map[string]int{
	"Morceau de moquette": 12,
	"Tuyau rouillé":       16,
	"Néon cassé":          14,
	"Barre de fer":        22,
}

// Buy — le joueur achète un article au marchand.
func (c *Character) Buy(item ShopItem) error {
	if item.IsSpellbook && c.KnowsSpell(item.Nom) {
		return ErrSpellAlreadyKnown
	}
	if c.Pieces < item.Prix {
		return ErrNotEnoughCoins
	}
	if !item.IsSpellbook && len(c.Inventaire) >= c.Capacite {
		return ErrInventoryFull
	}
	c.Pieces -= item.Prix
	if item.IsSpellbook {
		return c.LearnSpell(item.Nom) // erreur impossible ici (déjà vérifié)
	}
	c.AddItem(item.Nom)
	return nil
}

// Sell — le joueur revend un objet de loot au marchand, depuis
// l'emplacement i de son inventaire.
func (c *Character) Sell(i int) (int, error) {
	if i < 0 || i >= len(c.Inventaire) {
		return 0, ErrItemNotFound
	}
	name := c.Inventaire[i]
	prix, ok := MerchantBuyPrices[name]
	if !ok {
		return 0, ErrItemNotSellable
	}
	c.RemoveItem(i)
	c.Pieces += prix
	return prix, nil
}

// sellableSlots renvoie l'emplacement de chaque objet que le marchand
// rachète, dans l'ordre de l'inventaire (c'est l'ordre de l'onglet Vendre).
func (c *Character) sellableSlots() []int {
	var slots []int
	for i, name := range c.Inventaire {
		if _, ok := MerchantBuyPrices[name]; ok {
			slots = append(slots, i)
		}
	}
	return slots
}

// Recipe — recette du forgeron : matériaux + pièces contre une pièce
// d'équipement.
type Recipe struct {
	Result    EquipmentItem
	Materials map[string]int
	Prix      int
}

// Recipes — une recette par emplacement (tête, torse, pieds), à base
// des matériaux lâchés par les monstres.
var Recipes = []Recipe{
	{ItemCasqueTuyau, map[string]int{"Tuyau rouillé": 2}, 30},
	{ItemPlastronMoquette, map[string]int{"Morceau de moquette": 3}, 50},
	{ItemBottesFer, map[string]int{"Barre de fer": 2}, 25},
}

// Craft — fabrique une pièce d'équipement : vérifie les pièces puis
// les matériaux (erreurs personnalisées), les consomme, équipe la
// pièce et renvoie l'ancien objet à l'inventaire (échange).
func (c *Character) Craft(r Recipe) error {
	if c.Pieces < r.Prix {
		return ErrNotEnoughCoins
	}
	for mat, qty := range r.Materials {
		if countItem(c.Inventaire, mat) < qty {
			return ErrMissingMaterials
		}
	}
	c.Pieces -= r.Prix
	for mat, qty := range r.Materials {
		c.removeItems(mat, qty)
	}
	item := r.Result
	c.EquipItem(&item)
	return nil
}
