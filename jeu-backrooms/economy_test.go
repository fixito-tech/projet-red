package main

import (
	"errors"
	"testing"
)

// Tâche : erreur personnalisée "pas assez de pièces".
func TestBuyNotEnoughCoins(t *testing.T) {
	h := NewCharacter("Test", Classes[0])
	h.Pieces = 0
	if err := h.Buy(MerchantSells[0]); !errors.Is(err, ErrNotEnoughCoins) {
		t.Errorf("erreur = %v, attendu ErrNotEnoughCoins", err)
	}
}

// Tâche : erreur personnalisée "inventaire plein".
func TestBuyInventoryFull(t *testing.T) {
	h := NewCharacter("Test", Classes[0])
	h.Pieces = 10000
	for len(h.Inventaire) < h.Capacite {
		h.Inventaire = append(h.Inventaire, "Barre de fer")
	}
	if err := h.Buy(MerchantSells[0]); !errors.Is(err, ErrInventoryFull) {
		t.Errorf("erreur = %v, attendu ErrInventoryFull", err)
	}
}

// Acheter le grimoire l'apprend ; le racheter renvoie une erreur.
func TestBuySpellbookOnceOnly(t *testing.T) {
	h := NewCharacter("Test", Classes[0])
	h.Pieces = 10000
	var grimoire ShopItem
	for _, it := range MerchantSells {
		if it.IsSpellbook {
			grimoire = it
		}
	}
	if err := h.Buy(grimoire); err != nil {
		t.Fatalf("premier achat refusé : %v", err)
	}
	if !h.KnowsSpell(grimoire.Nom) {
		t.Fatal("le sort n'a pas été appris")
	}
	if err := h.Buy(grimoire); !errors.Is(err, ErrSpellAlreadyKnown) {
		t.Errorf("erreur = %v, attendu ErrSpellAlreadyKnown", err)
	}
}

// Vendre un objet qui n'est pas du loot renvoie une erreur, vendre du
// loot rapporte des pièces et le retire de l'inventaire.
func TestSellLootAndRejectOthers(t *testing.T) {
	h := NewCharacter("Test", Classes[0]) // possède 3x Almond Water
	if _, err := h.Sell(0); !errors.Is(err, ErrItemNotSellable) {
		t.Errorf("vendre une Almond Water devrait échouer, erreur = %v", err)
	}
	h.Inventaire = append(h.Inventaire, "Tuyau rouillé")
	avant := h.Pieces
	prix, err := h.Sell(len(h.Inventaire) - 1)
	if err != nil {
		t.Fatalf("vente refusée : %v", err)
	}
	if prix != MerchantBuyPrices["Tuyau rouillé"] || h.Pieces != avant+prix {
		t.Errorf("prix = %d, pièces = %d", prix, h.Pieces)
	}
	if countItem(h.Inventaire, "Tuyau rouillé") != 0 {
		t.Error("l'objet vendu est toujours dans l'inventaire")
	}
}

// Tâche : erreur personnalisée "matériaux manquants".
func TestCraftMissingMaterials(t *testing.T) {
	h := NewCharacter("Test", Classes[0])
	h.Pieces = 10000
	if err := h.Craft(Recipes[0]); !errors.Is(err, ErrMissingMaterials) {
		t.Errorf("erreur = %v, attendu ErrMissingMaterials", err)
	}
}

// Le forgeron fabrique, équipe, consomme matériaux + pièces.
func TestCraftSuccessEquipsAndConsumes(t *testing.T) {
	h := NewCharacter("Test", Classes[0])
	r := Recipes[0] // Casque en tuyau : 2x Tuyau rouillé, 30 po
	h.Pieces = 100
	h.Inventaire = append(h.Inventaire, "Tuyau rouillé", "Tuyau rouillé")

	if err := h.Craft(r); err != nil {
		t.Fatalf("fabrication refusée : %v", err)
	}
	if h.Pieces != 70 {
		t.Errorf("pièces restantes = %d, attendu 70", h.Pieces)
	}
	if countItem(h.Inventaire, "Tuyau rouillé") != 0 {
		t.Error("les matériaux n'ont pas été consommés")
	}
	if h.Equip.Get(r.Result.Slot) == nil || h.Equip.Get(r.Result.Slot).Nom != r.Result.Nom {
		t.Error("la pièce fabriquée n'a pas été équipée")
	}
}

func TestCraftNotEnoughCoins(t *testing.T) {
	h := NewCharacter("Test", Classes[0])
	h.Pieces = 0
	h.Inventaire = append(h.Inventaire, "Tuyau rouillé", "Tuyau rouillé")
	if err := h.Craft(Recipes[0]); !errors.Is(err, ErrNotEnoughCoins) {
		t.Errorf("erreur = %v, attendu ErrNotEnoughCoins", err)
	}
}
