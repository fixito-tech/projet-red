package main

// ---------------------------------------------------------------
// ÉQUIPEMENT — tâche « struct Equipment (tête, torse, pieds) »
// Logique pure : aucun affichage ici (voir economy_ui.go pour le
// forgeron et ui.go pour le HUD).
// ---------------------------------------------------------------

// EquipmentSlot identifie l'un des trois emplacements d'équipement.
type EquipmentSlot int

const (
	SlotHead EquipmentSlot = iota
	SlotChest
	SlotFeet
)

func (s EquipmentSlot) String() string {
	switch s {
	case SlotHead:
		return "Tête"
	case SlotChest:
		return "Torse"
	case SlotFeet:
		return "Pieds"
	}
	return "?"
}

// EquipmentItem — une pièce d'équipement fabriquée par le forgeron.
// BonusPV s'ajoute aux PV max du personnage tant que l'objet est porté.
type EquipmentItem struct {
	Nom     string
	Slot    EquipmentSlot
	BonusPV int
}

// Equipment — tâche : struct avec un emplacement par zone du corps.
type Equipment struct {
	Head, Chest, Feet *EquipmentItem
}

// slotPtr renvoie un pointeur vers le bon champ selon l'emplacement,
// pour éviter de dupliquer Set/Get/TotalBonusPV trois fois.
func (e *Equipment) slotPtr(slot EquipmentSlot) **EquipmentItem {
	switch slot {
	case SlotHead:
		return &e.Head
	case SlotChest:
		return &e.Chest
	default:
		return &e.Feet
	}
}

// Set équipe item et renvoie l'objet précédemment porté à cet
// emplacement (nil si l'emplacement était vide) : c'est l'« échange
// avec l'ancien objet » demandé par l'énoncé.
func (e *Equipment) Set(item *EquipmentItem) *EquipmentItem {
	p := e.slotPtr(item.Slot)
	old := *p
	*p = item
	return old
}

// Get renvoie l'objet porté à un emplacement, ou nil.
func (e Equipment) Get(slot EquipmentSlot) *EquipmentItem {
	switch slot {
	case SlotHead:
		return e.Head
	case SlotChest:
		return e.Chest
	default:
		return e.Feet
	}
}

// TotalBonusPV additionne les bonus de PV max des trois emplacements.
func (e Equipment) TotalBonusPV() int {
	total := 0
	for _, it := range []*EquipmentItem{e.Head, e.Chest, e.Feet} {
		if it != nil {
			total += it.BonusPV
		}
	}
	return total
}

// Count renvoie le nombre d'emplacements occupés (0 à 3) : c'est ce
// nombre qui réduit l'aura du boss (voir boss.go).
func (e Equipment) Count() int {
	n := 0
	for _, it := range []*EquipmentItem{e.Head, e.Chest, e.Feet} {
		if it != nil {
			n++
		}
	}
	return n
}

// Les trois pièces fabricables par le forgeron (voir economy.go pour
// les recettes). Une seule pièce par emplacement, donc 3 au total.
var (
	ItemCasqueTuyau      = EquipmentItem{"Casque en tuyau", SlotHead, 20}
	ItemPlastronMoquette = EquipmentItem{"Plastron de moquette", SlotChest, 35}
	ItemBottesFer        = EquipmentItem{"Bottes en fer", SlotFeet, 15}
)

// EquipmentCatalogue permet de retrouver la définition complète d'une
// pièce d'équipement à partir de son nom (ex. quand elle repasse dans
// l'inventaire après un échange, puis qu'on la réutilise depuis le sac).
var EquipmentCatalogue = map[string]EquipmentItem{
	ItemCasqueTuyau.Nom:      ItemCasqueTuyau,
	ItemPlastronMoquette.Nom: ItemPlastronMoquette,
	ItemBottesFer.Nom:        ItemBottesFer,
}
