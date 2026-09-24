package main

// ---------------------------------------------------------------
// ERREURS PERSONNALISÉES — utilisées par l'économie (marchand,
// forgeron) et les sorts. Toujours des phrases prêtes à afficher
// telles quelles dans un toast.
// ---------------------------------------------------------------

// GameError est une erreur simple portant directement son message
// utilisateur : pas besoin de la reformuler avant de l'afficher.
type GameError string

// Error renvoie le message de l'erreur, prêt à être affiché.
func (e GameError) Error() string { return string(e) }

const (
	ErrNotEnoughCoins    = GameError("Pas assez de pièces.")
	ErrInventoryFull     = GameError("Inventaire plein.")
	ErrMissingMaterials  = GameError("Matériaux manquants.")
	ErrSpellAlreadyKnown = GameError("Ce sort est déjà appris.")
	ErrItemNotSellable   = GameError("Le marchand n'achète pas cet objet.")
	ErrItemNotFound      = GameError("Objet introuvable dans l'inventaire.")
)
