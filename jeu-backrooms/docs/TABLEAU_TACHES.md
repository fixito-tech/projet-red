# Tableau de correspondance — tâches de l'énoncé

⚠️ **Le PDF de l'énoncé n'a pas été fourni pour cette implémentation.**
Le développement s'est basé sur la description détaillée transmise en
consigne (parties « personnage et inventaire », « économie et
artisanat », « combat au tour par tour », + missions bonus). Les
numéros de tâche ci-dessous ne sont donc **certains** que lorsqu'ils
étaient déjà présents dans le code fourni (commentaires `tâche N`
d'origine dans `player.go`) ; partout ailleurs, la colonne « # PDF »
indique `?` et doit être vérifiée par l'équipe contre le vrai PDF avant
la soutenance. Le regroupement en 3 parties + bonus, lui, suit
fidèlement la consigne.

**Vérification** : le projet n'a pas de tests automatisés. La colonne
« Vérif. » renvoie aux scénarios de la
[checklist manuelle](CHECKLIST_MANUELLE.md), à rejouer avant chaque
remise et après chaque refactoring.

## Partie 1 — Personnage et inventaire

| # PDF | Tâche | Fichier | Fonction / type | Vérif. |
|---|---|---|---|---|
| 1 | `struct Character` | [player.go](../src/player.go) | `Character` | S1 |
| 11 | Création : nom (lettres uniquement, formaté), 3 classes, départ à 50 % des PV | [player.go](../src/player.go), [menu_input.go](../src/menu_input.go), [menu_ui.go](../src/menu_ui.go) | `NewCharacter`, `FormatNom`, `updateCreate`, `drawClassStep` | S1 |
| 5 | Potion de vie (Almond Water, +50 PV) | [inventory.go](../src/inventory.go) | `UseItem` | S3 |
| ? | Potion de poison : dégâts dans le temps, PV affichés | [combat.go](../src/combat.go) | `UseBagItem`, `endRound` (poison) | S7 |
| ? | Mort puis résurrection à 50 % des PV | [player.go](../src/player.go), [combat_input.go](../src/combat_input.go) | `Character.Die`, `endCombat` (cas `ResultLose`) | S10 |
| ? | Livre de sort appris une seule fois | [spells.go](../src/spells.go), [economy.go](../src/economy.go) | `LearnSpell`, `Buy` | S4, S6 |
| 12 | Inventaire : 10 emplacements au départ | [inventory.go](../src/inventory.go) | `AddItem`, `Capacite` | S3 |
| 18 | Inventaire : +10 emplacements, 3 fois maximum | [inventory.go](../src/inventory.go) | `UpgradeInventory` | S3 |
| ? | Argent de départ | [player.go](../src/player.go) | `StartingCoins`, `NewCharacter` | S1 |

## Partie 2 — Économie et artisanat

| # PDF | Tâche | Fichier | Fonction / type | Vérif. |
|---|---|---|---|---|
| ? | Marchand : achète le loot, vend potions/grimoire | [economy.go](../src/economy.go), [economy_ui.go](../src/economy_ui.go) | `Buy`, `Sell`, `MerchantSells`, `MerchantBuyPrices` | S4 |
| ? | Forgeron : recettes (matériaux + pièces) | [economy.go](../src/economy.go) | `Craft`, `Recipes` | S9 |
| ? | `struct Equipment` (tête, torse, pieds), bonus PV max, échange avec l'ancien objet | [equipment.go](../src/equipment.go), [player.go](../src/player.go) | `Equipment`, `Equipment.Set`, `Character.EquipItem` | S9 |
| ? | Interaction avec un PNJ (touche dédiée + bulle d'aide) | [economy_input.go](../src/economy_input.go), [economy_ui.go](../src/economy_ui.go) | `nearbyNPC`, `openNPC`, `drawInteractionBubble` | S4, S9 |
| ? | Erreurs personnalisées (pièces, inventaire plein, matériaux, sort déjà connu…) | [errors.go](../src/errors.go) | `GameError`, `ErrNotEnoughCoins`, etc. | S4, S9 |

## Partie 3 — Combat au tour par tour

| # PDF | Tâche | Fichier | Fonction / type | Vérif. |
|---|---|---|---|---|
| ? | `struct Monster` | [monster.go](../src/monster.go) | `Monster`, `NewMonster` | S5 |
| ? | Pattern d'attaque : dégâts doublés tous les 3 tours | [combat.go](../src/combat.go) | `monsterAttack` | S6 |
| ? | Tour du joueur : attaquer / inventaire | [combat.go](../src/combat.go), [combat_input.go](../src/combat_input.go) | `Combat.Attack`, `Combat.UseBagItem`, `updateCombat` | S6, S7 |
| ? | Boucle de combat avec compteur de tours | [combat.go](../src/combat.go) | `Combat.Round`, `startRound`, `endRound` | S6 |
| ? | Monstres : apparition, IA (errance/poursuite), contact → combat, réapparition | [monster.go](../src/monster.go), [world.go](../src/world.go) | `SpawnMonster`, `Monster.Update`, `updateMonsterRespawn` | S5, S8 |
| ? | Loot : drop au sol (max 3 types), ramassage | [monster.go](../src/monster.go), [world.go](../src/world.go) | `RollLoot`, `Drop`, `updateDropPickup` | S8 |
| ? | Écran de combat façon Pokémon (transition, barres, boîte de texte) | [combat_ui.go](../src/combat_ui.go) | `drawCombat`, `drawCombatBar` | S6 |
| ? | Boss : placement, confirmation, imbattable sans équipement, écran de victoire | [boss.go](../src/boss.go), [combat_input.go](../src/combat_input.go) | `NewBoss`, `BossAuraDamage`, `updateBossConfirm`, `endCombat` | S11 |

## Missions bonus

La consigne annonce 6 missions bonus mais n'en nomme explicitement que
4 dans la description transmise ; les 2 autres sont donc absentes de
ce tableau faute d'information (à compléter contre le PDF).

| Mission bonus | Fichier | Fonction / type | Vérif. |
|---|---|---|---|
| Initiative (le plus rapide agit en premier) | [player.go](../src/player.go), [combat.go](../src/combat.go) | `Character.Vitesse`, `Monster.Vitesse`, `startRound` | S6 |
| Expérience (XP, niveaux) | [player.go](../src/player.go) | `Character.AddXP` | S12 |
| Sorts (grimoire verrouillé) | [spells.go](../src/spells.go) | `LearnSpell`, `AttackLocked` | S4, S6 |
| Mana / énergie (coût, régénération en et hors combat) | [combat.go](../src/combat.go), [world.go](../src/world.go) | `EnergyRegenPerTurn`, `updateOutOfCombatRegen` | S6 |

## Divergences connues à vérifier contre le PDF

- **Arborescence `/src` + `/docs`** : **résolu**. Le code, `go.mod`,
  les images (`assets/`) et les scripts Python sont dans `src/` ;
  `docs/` et le README restent à la racine. On lance le jeu depuis
  `src/` (`cd src && go run .`).
- **Numérotation précise 1 à 22** : seules les tâches 1, 5, 11, 12 et
  18 étaient déjà annotées dans le code d'origine ; toutes les autres
  correspondances (`?`) sont des hypothèses de bonne foi à recaler sur
  le vrai PDF.
- **Équilibrage assumé** (décision d'équipe après playtest, voir le
  README) : 9 monstres au lieu de 6, et vie d'un monstre normal à 1x
  les PV max du joueur au lieu de 2x. Les constantes concernées sont
  en haut de `monster.go` et commentées comme telles.
- **Pas de tests automatisés** : validation par la
  [checklist manuelle](CHECKLIST_MANUELLE.md).
